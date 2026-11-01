import { useEffect, useState } from "react";
import { toErrorMessage } from "../utils/errors";
import {
  pomodoroTimer,
  stopwatchTimer,
  type PomodoroPhase,
  type PomodoroState,
  type StopwatchState,
} from "../services/timers";

type TimerMode = "pomodoro" | "stopwatch";

interface TimerPageProps {
  mode: TimerMode;
}

function formatTime(milliseconds: number): string {
  const totalSeconds = Math.floor(milliseconds / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const padded = (value: number) => String(value).padStart(2, "0");
  return hours > 0
    ? `${padded(hours)}:${padded(minutes)}:${padded(seconds)}`
    : `${padded(minutes)}:${padded(seconds)}`;
}

function phaseLabel(phase: PomodoroPhase): string {
  switch (phase) {
    case "focus":
      return "Focus";
    case "short_break":
      return "Short break";
    case "long_break":
      return "Long break";
    default:
      return phase;
  }
}

export function TimerPage({ mode }: TimerPageProps) {
  const [pomodoroSnapshot, setPomodoroSnapshot] = useState<PomodoroState | null>(null);
  const [stopwatchSnapshot, setStopwatchSnapshot] = useState<StopwatchState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const isPomodoro = mode === "pomodoro";
  const activeSnapshot = isPomodoro ? pomodoroSnapshot : stopwatchSnapshot;
  const status = activeSnapshot?.status;

  useEffect(() => {
    let mounted = true;
    let timeout: ReturnType<typeof setTimeout>;

    const poll = async () => {
      try {
        if (isPomodoro) {
          const next = await pomodoroTimer.getState();
          if (mounted) setPomodoroSnapshot(next);
        } else {
          const next = await stopwatchTimer.getState();
          if (mounted) setStopwatchSnapshot(next);
        }
        if (mounted) setError(null);
      } catch (pollError) {
        if (mounted) setError(toErrorMessage(pollError));
      } finally {
        if (mounted) timeout = setTimeout(poll, 250);
      }
    };

    void poll();
    return () => {
      mounted = false;
      clearTimeout(timeout);
    };
  }, [isPomodoro]);

  async function perform(action: "start" | "pause" | "resume" | "reset" | "advance") {
    setBusy(true);
    setError(null);
    try {
      if (isPomodoro) {
        await pomodoroTimer[action]();
        setPomodoroSnapshot(await pomodoroTimer.getState());
      } else if (action !== "advance") {
        await stopwatchTimer[action]();
        setStopwatchSnapshot(await stopwatchTimer.getState());
      }
    } catch (actionError) {
      setError(toErrorMessage(actionError));
    } finally {
      setBusy(false);
    }
  }

  const displayedTime = isPomodoro
    ? pomodoroSnapshot
      ? formatTime(pomodoroSnapshot.remainingMillis)
      : "--:--"
    : stopwatchSnapshot
      ? formatTime(stopwatchSnapshot.elapsedMillis)
      : "--:--";

  return (
    <section className="mx-auto max-w-3xl">
      <h1 className="text-2xl font-semibold">
        {isPomodoro ? "Focus Timer" : "Stopwatch"}
      </h1>
      <p className="mt-1 text-slate-400">
        {isPomodoro
          ? "Work in focused sessions, with breaks between phases."
          : "Track elapsed study time with a simple stopwatch."}
      </p>

      <div className="mt-8 rounded-2xl border border-edge bg-panel px-6 py-10 text-center sm:px-12">
        {pomodoroSnapshot && (
          <p className="text-sm font-medium uppercase tracking-widest text-slate-400">
            {phaseLabel(pomodoroSnapshot.phase)}
          </p>
        )}
        <p
          className="mt-3 font-mono text-6xl font-semibold tabular-nums tracking-tight sm:text-8xl"
          aria-label={isPomodoro ? "Time remaining" : "Time elapsed"}
          aria-live="off"
        >
          {displayedTime}
        </p>
        <p className="mt-3 text-sm capitalize text-slate-400" role="status">
          {activeSnapshot ? status : "loading"}
        </p>
        {pomodoroSnapshot && (
          <p className="mt-1 text-sm text-slate-400">
            {pomodoroSnapshot.completedFocuses} focus sessions completed
          </p>
        )}

        <div className="mt-8 flex flex-wrap justify-center gap-3">
          {status === "running" ? (
            <ControlButton disabled={busy} onClick={() => void perform("pause")}>
              Pause
            </ControlButton>
          ) : status === "paused" ? (
            <ControlButton disabled={busy} onClick={() => void perform("resume")}>
              Resume
            </ControlButton>
          ) : status === "completed" && isPomodoro ? (
            <ControlButton disabled={busy} onClick={() => void perform("advance")}>
              Next phase
            </ControlButton>
          ) : (
            <ControlButton disabled={busy || !activeSnapshot} onClick={() => void perform("start")}>
              Start
            </ControlButton>
          )}
          <button
            type="button"
            disabled={busy || !activeSnapshot}
            onClick={() => void perform("reset")}
            className="rounded-md border border-edge px-5 py-2.5 text-sm font-medium text-slate-200 transition-colors hover:bg-edge/60 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Reset
          </button>
        </div>
      </div>
      {error && (
        <p className="mt-4 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
          {error}
        </p>
      )}
    </section>
  );
}

interface ControlButtonProps {
  children: string;
  disabled: boolean;
  onClick: () => void;
}

function ControlButton({ children, disabled, onClick }: ControlButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="rounded-md bg-accent px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
    >
      {children}
    </button>
  );
}
