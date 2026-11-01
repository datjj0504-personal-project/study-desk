import {
  AdvancePomodoro,
  GetPomodoroState,
  GetStopwatchState,
  PausePomodoro,
  PauseStopwatch,
  ResetPomodoro,
  ResetStopwatch,
  ResumePomodoro,
  ResumeStopwatch,
  StartPomodoro,
  StartStopwatch,
} from "../../wailsjs/go/main/App";

export type TimerStatus = string;
export type PomodoroPhase = string;

export interface PomodoroState {
  phase: PomodoroPhase;
  status: TimerStatus;
  durationMillis: number;
  elapsedMillis: number;
  remainingMillis: number;
  completedFocuses: number;
}

export interface StopwatchState {
  status: TimerStatus;
  elapsedMillis: number;
}

export const pomodoroTimer = {
  getState: (): Promise<PomodoroState> => GetPomodoroState(),
  start: (): Promise<void> => StartPomodoro(),
  pause: (): Promise<void> => PausePomodoro(),
  resume: (): Promise<void> => ResumePomodoro(),
  reset: (): Promise<void> => ResetPomodoro(),
  advance: (): Promise<void> => AdvancePomodoro(),
};

export const stopwatchTimer = {
  getState: (): Promise<StopwatchState> => GetStopwatchState(),
  start: (): Promise<void> => StartStopwatch(),
  pause: (): Promise<void> => PauseStopwatch(),
  resume: (): Promise<void> => ResumeStopwatch(),
  reset: (): Promise<void> => ResetStopwatch(),
};
