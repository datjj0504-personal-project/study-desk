import { afterEach, describe, expect, it, vi } from "vitest";
import { pomodoroTimer, stopwatchTimer } from "./timers";

const { bindings } = vi.hoisted(() => ({
  bindings: {
    AdvancePomodoro: vi.fn(),
    GetPomodoroState: vi.fn(),
    GetStopwatchState: vi.fn(),
    PausePomodoro: vi.fn(),
    PauseStopwatch: vi.fn(),
    ResetPomodoro: vi.fn(),
    ResetStopwatch: vi.fn(),
    ResumePomodoro: vi.fn(),
    ResumeStopwatch: vi.fn(),
    StartPomodoro: vi.fn(),
    StartStopwatch: vi.fn(),
  },
}));

vi.mock("../../wailsjs/go/main/App", () => bindings);

afterEach(() => vi.resetAllMocks());

describe("timer services", () => {
  it("exposes Pomodoro state and controls", async () => {
    const state = {
      phase: "focus",
      status: "running",
      durationMillis: 1_500_000,
      elapsedMillis: 1_000,
      remainingMillis: 1_499_000,
      completedFocuses: 2,
    };
    bindings.GetPomodoroState.mockResolvedValue(state);
    bindings.StartPomodoro.mockResolvedValue(undefined);
    bindings.PausePomodoro.mockResolvedValue(undefined);
    bindings.ResumePomodoro.mockResolvedValue(undefined);
    bindings.ResetPomodoro.mockResolvedValue(undefined);
    bindings.AdvancePomodoro.mockResolvedValue(undefined);

    await expect(pomodoroTimer.getState()).resolves.toEqual(state);
    await pomodoroTimer.start();
    await pomodoroTimer.pause();
    await pomodoroTimer.resume();
    await pomodoroTimer.reset();
    await pomodoroTimer.advance();

    expect(bindings.StartPomodoro).toHaveBeenCalledOnce();
    expect(bindings.PausePomodoro).toHaveBeenCalledOnce();
    expect(bindings.ResumePomodoro).toHaveBeenCalledOnce();
    expect(bindings.ResetPomodoro).toHaveBeenCalledOnce();
    expect(bindings.AdvancePomodoro).toHaveBeenCalledOnce();
  });

  it("exposes stopwatch state and controls", async () => {
    const state = { status: "paused", elapsedMillis: 12_345 };
    bindings.GetStopwatchState.mockResolvedValue(state);
    bindings.StartStopwatch.mockResolvedValue(undefined);
    bindings.PauseStopwatch.mockResolvedValue(undefined);
    bindings.ResumeStopwatch.mockResolvedValue(undefined);
    bindings.ResetStopwatch.mockResolvedValue(undefined);

    await expect(stopwatchTimer.getState()).resolves.toEqual(state);
    await stopwatchTimer.start();
    await stopwatchTimer.pause();
    await stopwatchTimer.resume();
    await stopwatchTimer.reset();

    expect(bindings.StartStopwatch).toHaveBeenCalledOnce();
    expect(bindings.PauseStopwatch).toHaveBeenCalledOnce();
    expect(bindings.ResumeStopwatch).toHaveBeenCalledOnce();
    expect(bindings.ResetStopwatch).toHaveBeenCalledOnce();
  });

  it("propagates backend errors", async () => {
    bindings.StartPomodoro.mockRejectedValue(new Error("timer unavailable"));

    await expect(pomodoroTimer.start()).rejects.toThrow("timer unavailable");
  });
});
