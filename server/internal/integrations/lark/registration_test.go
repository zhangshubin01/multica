package lark

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// registrationFake is a tiny HTTP stand-in for accounts.feishu.cn /
// accounts.larksuite.com. Each test registers a sequence of poll
// handlers and assertions about what the client sent.
type registrationFake struct {
	t   *testing.T
	srv *httptest.Server
	mux *http.ServeMux

	beginN atomic.Int32
	pollN  atomic.Int32
}

func newRegistrationFake(t *testing.T) *registrationFake {
	t.Helper()
	f := &registrationFake{t: t, mux: http.NewServeMux()}
	f.srv = httptest.NewServer(f.mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *registrationFake) URL() string { return f.srv.URL }

// stubBegin pins the /oauth/v1/app/registration response for action=begin.
// Verifies the client sent the canonical PersonalAgent registration form.
func (f *registrationFake) stubBegin(resp map[string]any) {
	f.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		switch r.Form.Get("action") {
		case "begin":
			f.beginN.Add(1)
			if got := r.Form.Get("archetype"); got != "PersonalAgent" {
				f.t.Errorf("begin: archetype=%q want PersonalAgent", got)
			}
			if got := r.Form.Get("auth_method"); got != "client_secret" {
				f.t.Errorf("begin: auth_method=%q want client_secret", got)
			}
			if got := r.Form.Get("request_user_info"); got != "open_id tenant_brand" {
				f.t.Errorf("begin: request_user_info=%q want %q", got, "open_id tenant_brand")
			}
			writeJSON(w, resp)
		case "poll":
			f.t.Errorf("poll called before stubBegin handler was replaced — tests should set stubPoll explicitly")
		default:
			f.t.Errorf("unknown action: %q", r.Form.Get("action"))
		}
	})
}

func TestRegistrationClient_Begin_HappyPath(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.stubBegin(map[string]any{
		"device_code": "dc_xyz",
		"user_code":   "ABCD-EFGH",
		"interval":    3,
		"expire_in":   600,
	})

	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Begin(context.Background(), "")
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if res.DeviceCode != "dc_xyz" {
		t.Errorf("DeviceCode: got %q", res.DeviceCode)
	}
	if res.Interval != 3*time.Second {
		t.Errorf("Interval: got %v want 3s", res.Interval)
	}
	if res.ExpiresIn != 600*time.Second {
		t.Errorf("ExpiresIn: got %v want 600s", res.ExpiresIn)
	}
	if res.Domain != fake.URL() {
		t.Errorf("Domain: got %q want %q", res.Domain, fake.URL())
	}
	// The QR URL must be the current official /page/cli flow on the
	// Feishu open host (empty region → Feishu default): Lark deprecated
	// the old verification_uri_complete (/page/launcher) target, so the
	// client builds this itself from the returned user_code.
	u, err := url.Parse(res.QRCodeURL)
	if err != nil {
		t.Fatalf("QRCodeURL: %v", err)
	}
	if u.Scheme != "https" || u.Host != "open.feishu.cn" || u.Path != "/page/cli" {
		t.Errorf("qr URL: got %q want https://open.feishu.cn/page/cli", res.QRCodeURL)
	}
	q := u.Query()
	if q.Get("user_code") != "ABCD-EFGH" {
		t.Errorf("qr user_code=%q want %q", q.Get("user_code"), "ABCD-EFGH")
	}
	if q.Get("from") != "cli" {
		t.Errorf("qr from=%q want cli", q.Get("from"))
	}
}

// TestRegistrationClient_Begin_MissingUserCodeIsInvalidResponse pins the
// new required field: without user_code the client cannot build the
// /page/cli URL, so Begin must fail as a typed protocol error instead
// of emitting a URL-less QR.
func TestRegistrationClient_Begin_MissingUserCodeIsInvalidResponse(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.stubBegin(map[string]any{
		"device_code": "dc_nousercode",
	})

	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	_, err := c.Begin(context.Background(), "")
	if err == nil {
		t.Fatal("want error on begin response without user_code")
	}
	var re *RegistrationError
	if !errorsAs(err, &re) || re.Code != "invalid_response" {
		t.Errorf("want invalid_response RegistrationError, got %v", err)
	}
}

// TestRegistrationClient_Begin_RegionLarkBeginsOnFeishu pins the
// post-launcher-deprecation routing: region=lark does NOT switch the
// begin endpoint — accounts.larksuite.com's begin still serves the
// legacy flow Lark international now rejects, so every begin (Lark
// included) must POST to the Feishu host and rely on the poll stream's
// tenant_brand="lark" switch (Poll) to flip hosts mid-session. What
// region=lark DOES change is the open host the verification URL is
// built on: open.larksuite.com/page/cli, so the user scans the cloud
// they will authorize from.
func TestRegistrationClient_Begin_RegionLarkBeginsOnFeishu(t *testing.T) {
	feishuFake := newRegistrationFake(t)
	feishuFake.stubBegin(map[string]any{
		"device_code": "dc_lark",
		"user_code":   "ABCD-EFGH",
	})

	larkFake := newRegistrationFake(t)
	larkFake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("begin must NOT POST to the Lark host (%s) — the official flow always starts on Feishu", larkFake.URL())
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := NewRegistrationClient(RegistrationConfig{
		Domain:     feishuFake.URL(),
		LarkDomain: larkFake.URL(),
	})
	res, err := c.Begin(context.Background(), RegionLark)
	if err != nil {
		t.Fatalf("Begin(region=lark): %v", err)
	}
	if res.Domain != feishuFake.URL() {
		t.Errorf("BeginResult.Domain: got %q want %q (Feishu) — subsequent polls start on Feishu and switch via tenant_brand",
			res.Domain, feishuFake.URL())
	}
	if got := feishuFake.beginN.Load(); got != 1 {
		t.Errorf("Feishu begin POSTs: got %d want 1", got)
	}
	if got := larkFake.beginN.Load(); got != 0 {
		t.Errorf("Lark begin POSTs: got %d want 0 (begin must never hit the Lark host)", got)
	}
	u, err := url.Parse(res.QRCodeURL)
	if err != nil {
		t.Fatalf("QRCodeURL: %v", err)
	}
	if u.Scheme != "https" || u.Host != "open.larksuite.com" || u.Path != "/page/cli" {
		t.Errorf("qr URL: got %q want https://open.larksuite.com/page/cli", res.QRCodeURL)
	}
	if u.Query().Get("from") != "cli" {
		t.Errorf("qr from=%q want cli", u.Query().Get("from"))
	}
}

// TestRegistrationClient_Begin_RegionFeishuBeginsOnFeishu pins the
// explicit-feishu side of the same split: passing region=feishu (or
// the empty-string back-compat default) keeps the original mainland
// host. Documenting both directions catches a future regression where
// the region selector accidentally inverts.
func TestRegistrationClient_Begin_RegionFeishuBeginsOnFeishu(t *testing.T) {
	feishuFake := newRegistrationFake(t)
	feishuFake.stubBegin(map[string]any{
		"device_code": "dc_feishu",
		"user_code":   "ABCD-EFGH",
	})

	larkFake := newRegistrationFake(t)
	larkFake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("region=feishu should NOT POST begin to the Lark host (%s)", larkFake.URL())
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := NewRegistrationClient(RegistrationConfig{
		Domain:     feishuFake.URL(),
		LarkDomain: larkFake.URL(),
	})
	res, err := c.Begin(context.Background(), RegionFeishu)
	if err != nil {
		t.Fatalf("Begin(region=feishu): %v", err)
	}
	if res.Domain != feishuFake.URL() {
		t.Errorf("BeginResult.Domain: got %q want %q (Feishu)", res.Domain, feishuFake.URL())
	}
	if got := feishuFake.beginN.Load(); got != 1 {
		t.Errorf("Feishu begin POSTs: got %d want 1", got)
	}
	if got := larkFake.beginN.Load(); got != 0 {
		t.Errorf("Lark begin POSTs: got %d want 0", got)
	}
	u, err := url.Parse(res.QRCodeURL)
	if err != nil {
		t.Fatalf("QRCodeURL: %v", err)
	}
	if u.Host != "open.feishu.cn" || u.Path != "/page/cli" {
		t.Errorf("qr URL: got %q want https://open.feishu.cn/page/cli", res.QRCodeURL)
	}
}

func TestRegistrationClient_Begin_HonorsExpiresInWireField(t *testing.T) {
	// MUL-7340: the wire field accounts.feishu.cn and
	// accounts.larksuite.com actually send is `expires_in` (RFC 8628
	// §3.2), currently 3600 — NOT the `expire_in` spelling the upstream
	// Go SDK types. Parsing only `expire_in` made every real begin
	// response look like "server omitted the expiry", silently
	// collapsing a 1-hour QR to the 10-minute fallback.
	fake := newRegistrationFake(t)
	fake.stubBegin(map[string]any{
		"device_code": "dc_expires_in",
		"user_code":   "ABCD-EFGH",
		"interval":    5,
		"expires_in":  3600,
	})

	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Begin(context.Background(), "")
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if res.ExpiresIn != 3600*time.Second {
		t.Errorf("ExpiresIn: got %v want 1h — the server's expires_in must win over the fallback", res.ExpiresIn)
	}
}

func TestRegistrationClient_Begin_DefaultsWhenServerOmitsTimers(t *testing.T) {
	// When Lark's response omits `interval` / `expires_in` (the empty
	// path the upstream SDK accepts), the client falls back to its
	// documented defaults rather than zero-second polling that would
	// hammer the endpoint or zero-second expiry that would fail the
	// very first poll.
	fake := newRegistrationFake(t)
	fake.stubBegin(map[string]any{
		"device_code": "dc_default",
		"user_code":   "ABCD-EFGH",
	})

	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Begin(context.Background(), "")
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if res.Interval != time.Duration(registrationDefaultPollSeconds)*time.Second {
		t.Errorf("interval default: got %v", res.Interval)
	}
	if res.ExpiresIn != time.Duration(registrationDefaultExpireSeconds)*time.Second {
		t.Errorf("expire default: got %v", res.ExpiresIn)
	}
}

func TestRegistrationClient_Begin_LarkError(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.stubBegin(map[string]any{
		"error":             "invalid_request",
		"error_description": "missing archetype",
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	_, err := c.Begin(context.Background(), "")
	if err == nil {
		t.Fatal("expected error from Lark error response")
	}
	var re *RegistrationError
	if !errorsAs(err, &re) {
		t.Fatalf("want *RegistrationError, got %T %v", err, err)
	}
	if re.Code != "invalid_request" {
		t.Errorf("Code: got %q want invalid_request", re.Code)
	}
}

func TestRegistrationClient_Begin_HTTPNon2xx(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("server boom"))
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	_, err := c.Begin(context.Background(), "")
	if err == nil {
		t.Fatal("want error on 500")
	}
	var re *RegistrationError
	if !errorsAs(err, &re) || re.Code != "http_500" {
		t.Errorf("want http_500 RegistrationError, got %v", err)
	}
}

func TestRegistrationClient_Poll_AuthorizationPending(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"error": "authorization_pending"})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Status != "authorization_pending" {
		t.Errorf("Status: got %q want authorization_pending", res.Status)
	}
	if res.ClientID != "" || res.SwitchedDomain != "" || res.Err != nil {
		t.Errorf("unexpected populated fields: %+v", res)
	}
}

// TestRegistrationClient_Poll_AuthorizationPendingHTTP400 pins the
// RFC 8628 transport behaviour Lark actually uses: the device-flow
// polling endpoint returns HTTP 400 with the JSON body
// `{"error":"authorization_pending"}` while the user hasn't scanned
// the QR yet. The previous transport treated any non-2xx as a
// terminal protocol error, which killed every install session on the
// first poll and made the QR dialog silently empty in the UI —
// because the frontend received status="error" + lark_protocol_error
// within seconds of opening the dialog.
//
// Verified against the live Lark service: it returns HTTP 400 with
// `code=20094` for the wait-state, not HTTP 200.
func TestRegistrationClient_Poll_AuthorizationPendingHTTP400(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"authorization_pending","error_description":"","code":20094}`))
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Status != "authorization_pending" {
		t.Errorf("Status: got %q want authorization_pending (HTTP 400 should NOT be a terminal error)", res.Status)
	}
	if res.Err != nil {
		t.Errorf("PollResult.Err should be nil for authorization_pending wait state, got %+v", res.Err)
	}
}

// TestRegistrationClient_Poll_AccessDeniedHTTP400 verifies the
// adjacent path: HTTP 400 with `error=access_denied` IS terminal (user
// cancelled in the Lark UI). The fix in doForm must distinguish these
// two cases by the body's `error` field, not by the HTTP status.
func TestRegistrationClient_Poll_AccessDeniedHTTP400(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"access_denied","error_description":"user cancelled"}`))
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Err == nil || res.Err.Code != "access_denied" {
		t.Errorf("PollResult.Err: got %+v; want access_denied", res.Err)
	}
}

// TestRegistrationClient_Poll_HTTP500UnparseableIsTerminal pins the
// fallback: a non-JSON 5xx (e.g. proxy returning a HTML error page) is
// still surfaced as a typed protocol error so ops can tell a Lark
// outage from a schema drift.
func TestRegistrationClient_Poll_HTTP500UnparseableIsTerminal(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html><body>502 Bad Gateway</body></html>"))
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	_, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err == nil {
		t.Fatal("want error on unparseable 502 body")
	}
	var re *RegistrationError
	if !errorsAs(err, &re) || re.Code != "http_502" {
		t.Errorf("want http_502 RegistrationError, got %v", err)
	}
}

func TestRegistrationClient_Poll_SlowDown(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"error": "slow_down"})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Status != "slow_down" {
		t.Errorf("Status: got %q want slow_down", res.Status)
	}
}

func TestRegistrationClient_Poll_EmptyBodyTreatedAsPending(t *testing.T) {
	// Lark sometimes returns an empty {}; treating it as "keep polling"
	// matches the upstream SDK and avoids a spurious abort during the
	// window between the user authorizing and the credentials being
	// minted on Lark's side.
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Status != "authorization_pending" {
		t.Errorf("Status: got %q want authorization_pending", res.Status)
	}
}

func TestRegistrationClient_Poll_DomainSwitchOnLarkTenant(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"user_info": map[string]any{"tenant_brand": "lark"},
		})
	})
	// LarkDomain points at a *distinct* host so we can assert the
	// switch surfaced the right value. We do NOT actually want the
	// client to hit it during this single Poll call — the switch
	// signal is consumed by RegistrationService, which re-polls
	// against the new host on the NEXT iteration.
	c := NewRegistrationClient(RegistrationConfig{
		Domain:     fake.URL(),
		LarkDomain: "https://lark-international.test",
	})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.SwitchedDomain != "https://lark-international.test" {
		t.Errorf("SwitchedDomain: got %q", res.SwitchedDomain)
	}
	if res.SwitchedRegion != RegionLark {
		t.Errorf("SwitchedRegion: got %q want %q", res.SwitchedRegion, RegionLark)
	}
}

// TestRegistrationClient_Poll_DomainSwitchOnFeishuTenant pins the
// reverse direction of the tenant-brand swap: a session begun against
// the Lark international host whose authorizing account turns out to
// be on mainland Feishu must surface a switch back to Feishu, with
// the region flipping accordingly. Without this, a user who picks the
// "Bind to Lark" CTA but actually scans with a Feishu account would
// carry RegionLark all the way through finishSuccess and either fail
// at GetBotInfo or commit a wrong-region installation row. Documenting
// this side keeps the symmetry promised in the public PollResult docs
// and the split-CTA UI's "wrong entry" recovery contract.
func TestRegistrationClient_Poll_DomainSwitchOnFeishuTenant(t *testing.T) {
	larkFake := newRegistrationFake(t)
	larkFake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"user_info": map[string]any{"tenant_brand": "feishu"},
		})
	})
	// Domain (Feishu) points at a *distinct* host so we can assert the
	// switch landed there; LarkDomain is the host we are CURRENTLY on
	// (the larkFake) so the swap predicate (`!HasPrefix(domain, LarkDomain)`)
	// resolves correctly.
	c := NewRegistrationClient(RegistrationConfig{
		Domain:     "https://feishu-mainland.test",
		LarkDomain: larkFake.URL(),
	})
	res, err := c.Poll(context.Background(), larkFake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.SwitchedDomain != "https://feishu-mainland.test" {
		t.Errorf("SwitchedDomain: got %q want feishu host", res.SwitchedDomain)
	}
	if res.SwitchedRegion != RegionFeishu {
		t.Errorf("SwitchedRegion: got %q want %q", res.SwitchedRegion, RegionFeishu)
	}
}

// TestRegistrationClient_Poll_NoSwitchWhenAlreadyOnMatchingHost pins
// that the swap is gated on the current domain — a `tenant_brand=lark`
// hint emitted while polling AGAINST the Lark host must NOT fire a
// redundant switch (which would loop the polling state machine), and
// likewise `tenant_brand=feishu` against the Feishu host. Both arms
// of the symmetry are covered to catch a future regression where the
// gate flips on only one side.
func TestRegistrationClient_Poll_NoSwitchWhenAlreadyOnMatchingHost(t *testing.T) {
	cases := []struct {
		name       string
		brand      string
		begunOn    string
		feishuHost string
		larkHost   string
	}{
		{
			name:       "lark brand on lark host is a no-op",
			brand:      "lark",
			begunOn:    "https://lark-international.test",
			feishuHost: "https://feishu-mainland.test",
			larkHost:   "https://lark-international.test",
		},
		{
			name:       "feishu brand on feishu host is a no-op",
			brand:      "feishu",
			begunOn:    "https://feishu-mainland.test",
			feishuHost: "https://feishu-mainland.test",
			larkHost:   "https://lark-international.test",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newRegistrationFake(t)
			fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, map[string]any{
					"user_info": map[string]any{"tenant_brand": tc.brand},
				})
			})
			// Use the fake's URL for whichever side we claim to be on,
			// and a placeholder URL for the other — the test only
			// exercises a single Poll call, the cross-host re-poll is
			// the service's job.
			cfg := RegistrationConfig{Domain: tc.feishuHost, LarkDomain: tc.larkHost}
			if tc.begunOn == tc.feishuHost {
				cfg.Domain = fake.URL()
			} else {
				cfg.LarkDomain = fake.URL()
			}
			c := NewRegistrationClient(cfg)
			res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
			if err != nil {
				t.Fatalf("Poll: %v", err)
			}
			if res.SwitchedDomain != "" {
				t.Errorf("SwitchedDomain: got %q, want empty (already on matching host)",
					res.SwitchedDomain)
			}
			if res.SwitchedRegion != "" {
				t.Errorf("SwitchedRegion: got %q, want empty", res.SwitchedRegion)
			}
		})
	}
}

func TestRegistrationClient_Poll_Success(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"client_id":     "cli_personal_42",
			"client_secret": "secret_42",
			"user_info":     map[string]any{"open_id": "ou_installer_42", "tenant_brand": "feishu"},
		})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.ClientID != "cli_personal_42" {
		t.Errorf("ClientID: got %q", res.ClientID)
	}
	if res.ClientSecret != "secret_42" {
		t.Errorf("ClientSecret: got %q", res.ClientSecret)
	}
	if string(res.OpenID) != "ou_installer_42" {
		t.Errorf("OpenID: got %q", res.OpenID)
	}
}

func TestRegistrationClient_Poll_SuccessMissingOpenIDIsProtocolError(t *testing.T) {
	// A "success" response without the installer open_id would leave
	// the auto-bind step with nothing to insert. Treat that as a hard
	// protocol error so RegistrationService never writes a half-built
	// installation row.
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"client_id":     "cli_personal_42",
			"client_secret": "secret_42",
		})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	_, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err == nil {
		t.Fatal("want error on success-without-open_id")
	}
	var re *RegistrationError
	if !errorsAs(err, &re) || re.Code != "invalid_response" {
		t.Errorf("want invalid_response, got %v", err)
	}
}

func TestRegistrationClient_Poll_AccessDenied(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"error":             "access_denied",
			"error_description": "user denied install",
		})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Err == nil || res.Err.Code != "access_denied" {
		t.Errorf("Err: %+v", res.Err)
	}
}

func TestRegistrationClient_Poll_ExpiredToken(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"error": "expired_token"})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Err == nil || res.Err.Code != "expired_token" {
		t.Errorf("Err: %+v", res.Err)
	}
}

func TestRegistrationClient_Poll_UnknownErrorPassesThrough(t *testing.T) {
	fake := newRegistrationFake(t)
	fake.mux.HandleFunc(registrationEndpoint, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"error": "rate_limited"})
	})
	c := NewRegistrationClient(RegistrationConfig{Domain: fake.URL()})
	res, err := c.Poll(context.Background(), fake.URL(), "dc_x")
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res.Err == nil || res.Err.Code != "rate_limited" {
		t.Errorf("Err: %+v", res.Err)
	}
}

func TestRegistrationClient_Poll_MissingDeviceCode(t *testing.T) {
	c := NewRegistrationClient(RegistrationConfig{})
	if _, err := c.Poll(context.Background(), "", ""); err == nil {
		t.Fatal("want error on missing device_code")
	}
}

// errorsAs is a tiny wrapper over errors.As so the test source stays
// terse — call sites read `errorsAs(err, &re)`.
func errorsAs(err error, target any) bool { return errors.As(err, target) }
