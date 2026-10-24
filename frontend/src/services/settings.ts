import {
  GetTimerSettings,
  UpdateTimerSettings,
} from "../../wailsjs/go/main/App";

export interface TimerSettings {
  focusMinutes: number;
  shortBreakMinutes: number;
  longBreakMinutes: number;
  longBreakEvery: number;
  notificationsEnabled: boolean;
}

export const timerSettings = {
  get: (): Promise<TimerSettings> => GetTimerSettings(),
  update: (settings: TimerSettings): Promise<void> =>
    UpdateTimerSettings(settings),
};
