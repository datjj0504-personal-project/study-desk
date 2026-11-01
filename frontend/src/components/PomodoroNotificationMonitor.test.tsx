import { act, cleanup, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PomodoroNotificationMonitor } from "./PomodoroNotificationMonitor";

const { getState } = vi.hoisted(() => ({
  getState: vi.fn(),
}));

vi.mock("../services/timers", () => ({
  pomodoroTimer: { getState },
}));

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.resetAllMocks();
});

describe("PomodoroNotificationMonitor", () => {
  it("notifies once when the timer transitions into completion", async () => {
    vi.useFakeTimers();
    getState
      .mockResolvedValueOnce({ phase: "focus", status: "running" })
      .mockResolvedValueOnce({ phase: "focus", status: "completed" })
      .mockResolvedValue({ phase: "focus", status: "completed" });
    const onCompletion = vi.fn();
    const onError = vi.fn();
    render(
      <PomodoroNotificationMonitor onCompletion={onCompletion} onError={onError} />,
    );

    await act(async () => {
      await Promise.resolve();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(onCompletion).toHaveBeenCalledOnce();
    expect(onCompletion).toHaveBeenCalledWith("focus");

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(onCompletion).toHaveBeenCalledOnce();
  });

  it("continues to detect completion independently of system notification preference", async () => {
    vi.useFakeTimers();
    getState
      .mockResolvedValueOnce({ phase: "short_break", status: "running" })
      .mockResolvedValueOnce({ phase: "short_break", status: "completed" });
    const onCompletion = vi.fn();
    render(
      <PomodoroNotificationMonitor onCompletion={onCompletion} onError={vi.fn()} />,
    );

    await act(async () => {
      await Promise.resolve();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(onCompletion).toHaveBeenCalledWith("short_break");
  });
});
