import { afterEach, describe, expect, it, vi } from "vitest";
import { listStudySessions } from "./sessions";

const { list } = vi.hoisted(() => ({ list: vi.fn() }));

vi.mock("../../wailsjs/go/main/App", () => ({ ListStudySessions: list }));
vi.mock("../../wailsjs/go/models", () => ({
  study: {
    Session: class {
      constructor(public source: Record<string, unknown>) {}
    },
  },
}));

afterEach(() => vi.resetAllMocks());

describe("study sessions service", () => {
  it("maps persisted sessions and optional subject IDs", async () => {
    list.mockResolvedValue([
      {
        id: 7,
        subjectId: 3,
        kind: "pomodoro",
        startedAt: "2026-10-24T08:00:00Z",
        endedAt: "2026-10-24T08:25:00Z",
        durationSeconds: 1500,
        note: "Read chapter 2",
      },
      {
        id: 6,
        subjectId: null,
        kind: "stopwatch",
        startedAt: "2026-10-23T08:00:00Z",
        endedAt: "2026-10-23T08:10:00Z",
        durationSeconds: 600,
        note: "",
      },
    ]);

    await expect(listStudySessions()).resolves.toEqual([
      {
        id: 7,
        subjectId: 3,
        kind: "pomodoro",
        startedAt: "2026-10-24T08:00:00Z",
        endedAt: "2026-10-24T08:25:00Z",
        durationSeconds: 1500,
        note: "Read chapter 2",
      },
      {
        id: 6,
        kind: "stopwatch",
        startedAt: "2026-10-23T08:00:00Z",
        endedAt: "2026-10-23T08:10:00Z",
        durationSeconds: 600,
        note: "",
      },
    ]);
  });

  it("surfaces unsupported session kinds and binding errors", async () => {
    list.mockResolvedValueOnce([{
      id: 1,
      subjectId: null,
      kind: "unknown",
      startedAt: "2026-10-24T08:00:00Z",
      endedAt: "2026-10-24T08:01:00Z",
      durationSeconds: 60,
      note: "",
    }]);
    await expect(listStudySessions()).rejects.toThrow("Unsupported study session kind: unknown");

    list.mockRejectedValueOnce(new Error("database unavailable"));
    await expect(listStudySessions()).rejects.toThrow("database unavailable");
  });
});
