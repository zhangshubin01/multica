import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// The composer hydrates every recent context entry it keeps by id on mount. An
// entry whose entity was deleted while this client was not listening would
// otherwise answer 404 on every mount, forever: the store only forgets a delete
// the client itself observed. Pin that a 404 evicts the entry and that any
// other failure keeps it (a 5xx must not erase a live recent item).
vi.mock("../../navigation", () => ({
  useNavigation: () => ({ pathname: "/ws-1/inbox", searchParams: new URLSearchParams() }),
}));

vi.mock("@multica/core/issues/queries", () => ({
  issueDetailOptions: (wsId: string, id: string) => ({
    queryKey: ["issue-detail", wsId, id],
    queryFn: async () => {
      const { ApiError } = await import("@multica/core/api");
      throw new ApiError("not found", 404, "Not Found");
    },
  }),
}));

vi.mock("@multica/core/projects/queries", () => ({
  projectDetailOptions: (wsId: string, id: string) => ({
    queryKey: ["project-detail", wsId, id],
    queryFn: async () => {
      const { ApiError } = await import("@multica/core/api");
      throw new ApiError("boom", 500, "Internal Server Error");
    },
  }),
}));

import { useRecentContextStore } from "@multica/core/chat";
import { useChatContextItems } from "./use-chat-context-items";

function recentIds(wsId: string): string[] {
  return (useRecentContextStore.getState().byWorkspace[wsId] ?? []).map((entry) => entry.id);
}

describe("useChatContextItems recent-context eviction", () => {
  beforeEach(() => {
    useRecentContextStore.setState({ byWorkspace: {} });
  });

  it("forgets a recent entry the server reports as 404 and keeps the rest", async () => {
    const { recordVisit } = useRecentContextStore.getState();
    recordVisit("ws-1", { type: "issue", id: "issue-deleted" });
    recordVisit("ws-1", { type: "project", id: "project-server-error" });
    expect(recentIds("ws-1")).toEqual(["project-server-error", "issue-deleted"]);

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );

    renderHook(() => useChatContextItems("ws-1"), { wrapper });

    await waitFor(() => expect(recentIds("ws-1")).toEqual(["project-server-error"]));
  });
});
