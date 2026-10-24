import {
  GetTheme,
  GetTimerSettings,
  UpdateTheme,
  UpdateTimerSettings,
} from "../../wailsjs/go/main/App";

export interface TimerSettings {
  focusMinutes: number;
  shortBreakMinutes: number;
  longBreakMinutes: number;
  longBreakEvery: number;
  notificationsEnabled: boolean;
}

export type Theme = "system" | "light" | "dark";

export const timerSettings = {
  get: (): Promise<TimerSettings> => GetTimerSettings(),
  update: (settings: TimerSettings): Promise<void> =>
    UpdateTimerSettings(settings),
};

export const themeSettings = {
  get: (): Promise<Theme> => GetTheme(),
  update: (theme: Theme): Promise<void> => UpdateTheme(theme),
};
