#!/usr/bin/env bash
# deploy/prod/bootstrap.sh — bring this checkout up as the production self-hosted
# stack on a fresh host, without hand-editing .env.
#
#   ./deploy/prod/bootstrap.sh                    # render .env, build, start, verify
#   ./deploy/prod/bootstrap.sh --render-only      # only render .env (no docker)
#   ./deploy/prod/bootstrap.sh --dry-run          # print the rendered .env, write nothing
#   ./deploy/prod/bootstrap.sh --extract-secrets  # carry this host's secrets to another
#
# Where things live
#   deploy/prod/env.template  every key the stack reads: @SECRET:NAME@ markers for
#                             secrets, @URL@/@WS@/@CIDR@ for host-derived values,
#                             @DBURL@ for the connection string built from POSTGRES_*
#   $SECRETS_FILE             NAME=value lines for the @SECRET: markers, mode 600,
#                             default ~/.multica-secrets/secrets.env — outside the repo
#
# An existing .env is never overwritten without --force, and then it is backed up.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE="$ROOT/deploy/prod/env.template"
TARGET="$ROOT/.env"
SECRETS_FILE="${SECRETS_FILE:-$HOME/.multica-secrets/secrets.env}"
COMPOSE=(docker compose
  -f docker-compose.selfhost.yml
  -f docker-compose.selfhost.build.yml
  -f docker-compose.lan.yml)

MODE=deploy
FORCE=0
BUILD=1
VERIFY=1
ALLOW_MISSING=0
HOST_IP=""
APP_URL=""

# A fresh host may mint these itself. Everything else stands for an account or an
# integration that already exists, so it has to be carried over.
GENERATABLE="POSTGRES_PASSWORD JWT_SECRET"

die() { printf 'error: %s\n' "$*" >&2; exit 1; }
note() { printf '  %s\n' "$*"; }
say() { printf '\n== %s\n' "$*"; }

gen_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c 64
  fi
}

# bash 3.2 (the system bash on macOS) has no associative arrays, so a secret is
# a lookup rather than a map entry. "Present but empty" is a deliberate value
# (the integration is unused); only a key that is absent counts as missing.
secret_present() {
  [ -f "$SECRETS_FILE" ] || return 1
  grep -qE "^$1=" "$SECRETS_FILE" 2>/dev/null
}

secret_value() {
  [ -f "$SECRETS_FILE" ] || return 0
  grep -E "^$1=" "$SECRETS_FILE" 2>/dev/null | head -1 | cut -d= -f2- || true
}

usage() {
  sed -n '2,18p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  cat <<'EOF'

Options
  --host <ip|host>          address to advertise (default: this host's LAN address)
  --app-url <url>           public/app URL, default http://<host>
  --secrets <file>          secrets file, default ~/.multica-secrets/secrets.env
  --target <file>           where to write .env, default <repo>/.env
  --force                   overwrite an existing .env (the old one is backed up)
  --render-only             render .env and stop before docker
  --dry-run                 print the rendered .env, write nothing
  --extract-secrets         write this checkout's secret values to the secrets file
  --allow-missing-secrets   blank missing secrets instead of failing
  --no-build                docker compose up without --build
  --no-verify               skip the post-start checks
  -h | --help               this text
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --host) HOST_IP="${2:?--host needs a value}"; shift 2 ;;
    --app-url) APP_URL="${2:?--app-url needs a value}"; shift 2 ;;
    --secrets) SECRETS_FILE="${2:?--secrets needs a value}"; shift 2 ;;
    --target) TARGET="${2:?--target needs a value}"; shift 2 ;;
    --force) FORCE=1; shift ;;
    --render-only) MODE=render; shift ;;
    --dry-run) MODE=dry; shift ;;
    --extract-secrets) MODE=extract; shift ;;
    --allow-missing-secrets) ALLOW_MISSING=1; shift ;;
    --no-build) BUILD=0; shift ;;
    --no-verify) VERIFY=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown argument: $1 (try --help)" ;;
  esac
done

[ -f "$TEMPLATE" ] || die "missing $TEMPLATE"

# ---------------------------------------------------------------- extract mode
if [ "$MODE" = extract ]; then
  [ -f "$ROOT/.env" ] || die "no .env in $ROOT to read secrets from"
  mkdir -p "$(dirname "$SECRETS_FILE")"
  chmod 700 "$(dirname "$SECRETS_FILE")" 2>/dev/null || true
  : > "$SECRETS_FILE"
  chmod 600 "$SECRETS_FILE"
  empty=()
  (
    umask 077
    while IFS= read -r line; do
      case "$line" in
        *=@SECRET:*)
          name="${line#*=@SECRET:}"; name="${name%@}"
          value="$(grep -E "^${name}=." "$ROOT/.env" | head -1 | cut -d= -f2- || true)"
          if [ -n "$value" ]; then
            printf '%s=%s\n' "$name" "$value"
          else
            printf '# %s is empty on the source host (that integration is unused there).\n' "$name"
            printf '# Keep it empty unless the target host uses it.\n'
            printf '%s=\n' "$name"
          fi
          ;;
      esac
    done < "$TEMPLATE"
  ) >> "$SECRETS_FILE"
  chmod 600 "$SECRETS_FILE"
  extracted="$(grep -c '^[A-Za-z_]' "$SECRETS_FILE" || true)"
  say "secrets extracted"
  note "file:   $SECRETS_FILE  ($extracted keys, mode 600, outside the repo)"
  note "carry:  scp \"$SECRETS_FILE\" mini:~/.multica-secrets/secrets.env"
  note "values were not printed, and nothing was written into the repository"
  exit 0
fi

# ------------------------------------------------------------- host detection
if [ -z "$HOST_IP" ]; then
  if command -v ipconfig >/dev/null 2>&1; then
    HOST_IP="$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || true)"
  fi
  if [ -z "$HOST_IP" ] && command -v hostname >/dev/null 2>&1; then
    HOST_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
  fi
fi
[ -n "$HOST_IP" ] || die "could not detect this host's address; pass --host <ip>"

[ -n "$APP_URL" ] || APP_URL="http://$HOST_IP"
APP_URL="${APP_URL%/}"
case "$APP_URL" in
  https://*) WS_URL="wss://${APP_URL#https://}" ;;
  http://*)  WS_URL="ws://${APP_URL#http://}" ;;
  *) die "--app-url must start with http:// or https://" ;;
esac
CIDR="$(printf '%s' "$HOST_IP" | awk -F. 'NF==4 {print $1"."$2"."$3".0/24"}' || true)"
[ -n "$CIDR" ] || CIDR="$HOST_IP/24"

# ---------------------------------------------------------------------- render
if [ ! -f "$SECRETS_FILE" ]; then
  note "no secrets file at $SECRETS_FILE"
fi

tmp="$(mktemp)"; trap 'rm -f "$tmp" "$tmp2"' EXIT; tmp2="$(mktemp)"
missing=(); generated=()
PG_USER=multica; PG_DB=multica; PG_PORT=5432; PG_PASS=""

# Pass 1 — every marker except @DBURL@, which needs POSTGRES_* from this pass.
while IFS= read -r line; do
  case "$line" in
    ''|\#*) printf '%s\n' "$line" >> "$tmp"; continue ;;
  esac
  key="${line%%=*}"; value="${line#*=}"
  case "$value" in
    @URL@)   value="$APP_URL" ;;
    @WS@)    value="$WS_URL" ;;
    @CIDR@)  value="$CIDR" ;;
    @EMPTY@) value="" ;;
    @SECRET:*)
      name="${value#@SECRET:}"; name="${name%@}"
      if secret_present "$name"; then
        value="$(secret_value "$name")"
      else
        value=""
        case " $GENERATABLE " in
          *" $name "*) value="$(gen_secret)"; generated+=("$name") ;;
          *) missing+=("$name") ;;
        esac
      fi
      ;;
  esac
  case "$key" in
    POSTGRES_USER) [ -n "$value" ] && PG_USER="$value" ;;
    POSTGRES_DB)   [ -n "$value" ] && PG_DB="$value" ;;
    POSTGRES_PORT) [ -n "$value" ] && PG_PORT="$value" ;;
    POSTGRES_PASSWORD) PG_PASS="$value" ;;
  esac
  printf '%s=%s\n' "$key" "$value" >> "$tmp"
done < "$TEMPLATE"

if [ ${#missing[@]} -gt 0 ]; then
  if [ "$ALLOW_MISSING" -eq 1 ]; then
    say "warning: these secrets stay empty"
    for n in "${missing[@]}"; do note "$n"; done
  else
    say "missing secrets (put them in $SECRETS_FILE or pass --secrets)"
    for n in "${missing[@]}"; do note "$n"; done
    note "a fresh host does not need them all: start with --allow-missing-secrets"
    die "${#missing[@]} secret(s) missing"
  fi
fi

# Pass 2 — DATABASE_URL from the values pass 1 resolved.
while IFS= read -r line; do
  case "$line" in
    DATABASE_URL=@DBURL@)
      printf 'DATABASE_URL=%s\n' \
        "postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" >> "$tmp2" ;;
    *) printf '%s\n' "$line" >> "$tmp2" ;;
  esac
done < "$tmp"

if grep -nE '^[A-Za-z_][A-Za-z0-9_]*=@' "$tmp2" >/dev/null 2>&1; then
  say "unresolved markers"
  grep -nE '^[A-Za-z_][A-Za-z0-9_]*=@' "$tmp2" >&2 || true
  die "the template still holds markers this script does not know"
fi

say "rendered"
note "host:   $HOST_IP   app: $APP_URL   ws: $WS_URL   cidr: $CIDR"
note "secrets file: $SECRETS_FILE$([ -f "$SECRETS_FILE" ] || printf ' (absent)')"
[ ${#generated[@]} -eq 0 ] || note "minted for this host: ${generated[*]}"
note "keys:   $(grep -cE '^[A-Za-z_][A-Za-z0-9_]*=' "$tmp2")"

if [ "$MODE" = dry ]; then
  printf '\n'
  cat "$tmp2"
  exit 0
fi

if [ -e "$TARGET" ] && [ "$FORCE" -ne 1 ]; then
  die "$TARGET exists — pass --force to replace it (it will be backed up)"
fi
if [ -e "$TARGET" ]; then
  backup="$TARGET.bak.$(date +%Y%m%d-%H%M%S)"
  cp -p "$TARGET" "$backup"
  note "previous .env saved as $(basename "$backup")"
fi
( umask 077; cat "$tmp2" > "$TARGET" )
chmod 600 "$TARGET"
note "wrote $TARGET (mode 600)"

[ "$MODE" = render ] && exit 0

# --------------------------------------------------------------------- deploy
say "starting the stack"
cd "$ROOT"
up_args=(up -d)
[ "$BUILD" -eq 1 ] && up_args+=(--build)
"${COMPOSE[@]}" "${up_args[@]}"

if [ "$VERIFY" -eq 1 ]; then
  say "waiting for the backend to report healthy"
  ok=0
  for _ in $(seq 1 60); do
    if curl -fsS --noproxy '*' http://127.0.0.1:8080/healthz 2>/dev/null | grep -q '"status":"ok"'; then
      ok=1; break
    fi
    sleep 3
  done
  [ "$ok" -eq 1 ] || die "backend did not become healthy in 180 s — docker logs multica-backend-1"

  say "verification"
  note "healthz:      $(curl -fsS --noproxy '*' http://127.0.0.1:8080/healthz)"
  note "migrations:   $(docker exec multica-postgres-1 psql -U "${PG_USER}" -d "${PG_DB}" -tAc 'select max(version) from schema_migrations' 2>/dev/null || echo 'unavailable')"
  note "flags:        $(curl -fsS --noproxy '*' http://127.0.0.1:8080/api/config 2>/dev/null | tr -d '\n' | grep -o '"docs_v1":[a-z]*' || echo '?') $(curl -fsS --noproxy '*' http://127.0.0.1:8080/api/config 2>/dev/null | tr -d '\n' | grep -o '"plugins_v1":[a-z]*' || true)"
  note "web (:3100):  $(curl -s --noproxy '*' -o /dev/null -w '%{http_code}' http://127.0.0.1:3100/)"
  note "edge (:80):   $(curl -s --noproxy '*' -o /dev/null -w '%{http_code}' http://127.0.0.1/)"
  "${COMPOSE[@]}" ps --format '{{.Name}} {{.Status}}' 2>/dev/null | while IFS= read -r l; do note "$l"; done
fi

say "done"
note "app:  $APP_URL"
note "logs: docker logs -f multica-backend-1"
note "host daemon still needs its own identity once: multica daemon start (PAT + runtime logins)"
