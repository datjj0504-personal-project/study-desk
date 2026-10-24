import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { TimerPage } from "./TimerPage";

const { api } = vi.hoisted(() => ({
  api: {
    pomodoroTimer: {
      advance: vi.fn(),
      getState: vi.fn(),
      pause: vi.fn(),
      reset: vi.fn(),
      resume: vi.fn(),
      start: vi.fn(),
    },
    stopwatchTimer: {
      getState: vi.fn(),
      pause: vi.fn(),
      reset: vi.fn(),
      resume: vi.fn(),
      start: vi.fn(),
    },
  },
}));

vi.mock("../services/timers", () => api);

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("TimerPage", () => {
  it("shows and controls the Pomodoro countdown", async () => {
    api.pomodoroTimer.getState
      .mockResolvedValueOnce({
        phase: "focus",
        status: "idle",
        durationMillis: 1_500_000,
        elapsedMillis: 0,
        remainingMillis: 1_500_000,
        completedFocuses: 0,
      })
      .mockResolvedValue({
        phase: "focus",
        status: "running",
        durationMillis: 1_500_000,
        elapsedMillis: 1_000,
        remainingMillis: 1_499_000,
        completedFocuses: 0,
      });
    api.pomodoroTimer.start.mockResolvedValue(undefined);
    api.pomodoroTimer.pause.mockResolvedValue(undefined);

    render(<TimerPage mode="pomodoro" />);
    expect(await screen.findByText("25:00")).toBeTruthy();
    expect(screen.getByText("Focus")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Start" }));
    expect(await screen.findByRole("button", { name: "Pause" })).toBeTruthy();
    expect(api.pomodoroTimer.start).toHaveBeenCalledOnce();

    fireEvent.click(screen.getByRole("button", { name: "Pause" }));
    await waitFor(() => expect(api.pomodoroTimer.pause).toHaveBeenCalledOnce());
  });

  it("shows stopwatch elapsed time and surfaces control failures", async () => {
    api.stopwatchTimer.getState.mockResolvedValue({
      status: "idle",
      elapsedMillis: 3_723_000,
    });
    api.stopwatchTimer.start.mockRejectedValue(new Error("timer unavailable"));

    render(<TimerPage mode="stopwatch" />);
    expect(await screen.findByText("01:02:03")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Start" }));
    expect((await screen.findByRole("alert")).textContent).toContain("timer unavailable");
  });

  it("advances a completed Pomodoro phase", async () => {
    api.pomodoroTimer.getState.mockResolvedValue({
      phase: "focus",
      status: "completed",
      durationMillis: 1_500_000,
      elapsedMillis: 1_500_000,
      remainingMillis: 0,
      completedFocuses: 1,
    });
    api.pomodoroTimer.advance.mockResolvedValue(undefined);

    render(<TimerPage mode="pomodoro" />);
    expect(await screen.findByText("00:00")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Next phase" }));

    await waitFor(() => expect(api.pomodoroTimer.advance).toHaveBeenCalledOnce());
  });
});
