import { useState } from "react";
import {
  themeSettings,
  timerSettings,
  type Theme,
  type TimerSettings,
} from "../services/settings";
import { toErrorMessage } from "../utils/errors";

interface SettingsPageProps {
  settings: TimerSettings | null;
  onSettingsChange: (settings: TimerSettings) => void;
  theme: Theme | null;
  onThemeChange: (theme: Theme) => void;
}

type DurationKey = "focusMinutes" | "shortBreakMinutes" | "longBreakMinutes" | "longBreakEvery";

const durationFields: Array<{
  key: DurationKey;
  label: string;
  min: number;
  max: number;
}> = [
  { key: "focusMinutes", label: "Focus duration (minutes)", min: 1, max: 180 },
  { key: "shortBreakMinutes", label: "Short break (minutes)", min: 1, max: 60 },
  { key: "longBreakMinutes", label: "Long break (minutes)", min: 1, max: 60 },
  { key: "longBreakEvery", label: "Long break after focus sessions", min: 1, max: 12 },
];

export function SettingsPage({
  settings,
  onSettingsChange,
  theme,
  onThemeChange,
}: SettingsPageProps) {
  const [draft, setDraft] = useState<TimerSettings | null>(null);
  const [themeDraft, setThemeDraft] = useState<Theme | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const currentDraft = draft ?? settings;
  const currentTheme = themeDraft ?? theme;

  async function save(value: TimerSettings) {
    setBusy(true);
    setError(null);
    setSaved(false);
    try {
      await timerSettings.update(value);
      setDraft(value);
      onSettingsChange(value);
      setSaved(true);
    } catch (saveError) {
      setError(toErrorMessage(saveError));
    } finally {
      setBusy(false);
    }
  }

  async function saveTheme(value: Theme) {
    setBusy(true);
    setError(null);
    setSaved(false);
    try {
      await themeSettings.update(value);
      setThemeDraft(value);
      onThemeChange(value);
      setSaved(true);
    } catch (saveError) {
      setError(toErrorMessage(saveError));
    } finally {
      setBusy(false);
    }
  }

  async function toggleNotifications() {
    if (!settings) return;
    if (settings.notificationsEnabled) {
      await save({ ...settings, notificationsEnabled: false });
      return;
    }
    if (typeof Notification === "undefined") {
      setError("System notifications are not available in this environment.");
      return;
    }
    try {
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        setError(permission === "denied"
          ? "Notification permission was denied. Change it in your system settings to enable alerts."
          : "Notification permission was not granted.");
        return;
      }
      await save({ ...settings, notificationsEnabled: true });
    } catch (permissionError) {
      setError(toErrorMessage(permissionError));
    }
  }

  return (
    <section className="mx-auto max-w-3xl">
      <h1 className="text-2xl font-semibold">Settings</h1>
      <p className="mt-1 text-slate-400">Configure your Pomodoro durations, completion alerts, and appearance.</p>

      {!currentDraft || !currentTheme ? (
        <p className="mt-6 text-slate-400" role="status">Loading settings…</p>
      ) : (
        <>
          <form
            className="mt-8 rounded-2xl border border-edge bg-panel p-6"
            onSubmit={(event) => {
              event.preventDefault();
              if (currentTheme) void saveTheme(currentTheme);
            }}
          >
            <h2 className="text-lg font-semibold">Appearance</h2>
            <p className="mt-2 text-sm text-slate-400">
              Choose a theme for StudyDesk. System follows your operating system preference.
            </p>
            <label className="mt-5 flex max-w-sm flex-col gap-2 text-sm text-slate-300">
              Theme
              <select
                value={currentTheme}
                onChange={(event) => {
                  const value = event.target.value;
                  if (value === "system" || value === "light" || value === "dark") {
                    setThemeDraft(value);
                    setSaved(false);
                  } else {
                    setError(`Unsupported theme: ${value}`);
                  }
                }}
                className="rounded-md border border-edge bg-surface px-3 py-2 text-slate-100 outline-none focus:border-accent"
              >
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </label>
            <button
              type="submit"
              disabled={busy || currentTheme === theme}
              className="mt-5 rounded-md bg-accent px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Save theme
            </button>
          </form>

          <form
            className="mt-5 rounded-2xl border border-edge bg-panel p-6"
            onSubmit={(event) => {
              event.preventDefault();
              if (currentDraft) void save(currentDraft);
            }}
          >
            <h2 className="text-lg font-semibold">Pomodoro durations</h2>
            <div className="mt-5 grid gap-4 sm:grid-cols-2">
              {durationFields.map(({ key, label, min, max }) => (
                <label key={key} className="flex flex-col gap-2 text-sm text-slate-300">
                  {label}
                  <input
                    type="number"
                    min={min}
                    max={max}
                    step={1}
                    required
                    value={currentDraft[key]}
                    onChange={(event) => {
                      const value = Number(event.target.value);
                      setDraft({ ...currentDraft, [key]: value });
                      setSaved(false);
                    }}
                    className="rounded-md border border-edge bg-surface px-3 py-2 text-slate-100 outline-none focus:border-accent"
                  />
                </label>
              ))}
            </div>
            <p className="mt-4 text-xs text-slate-400">Duration changes apply when the timer is idle. Defaults are 25 / 5 / 15 minutes, with a long break every 4 focus sessions.</p>
            <button
              type="submit"
              disabled={busy}
              className="mt-5 rounded-md bg-accent px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {busy ? "Saving…" : "Save durations"}
            </button>
          </form>

          <div className="mt-5 rounded-2xl border border-edge bg-panel p-6">
            <h2 className="text-lg font-semibold">Phase completion notifications</h2>
            <p className="mt-2 text-sm text-slate-400">
              Show an in-app alert when a Pomodoro phase ends. System notifications are optional and require permission.
            </p>
            <button
              type="button"
              disabled={busy || !settings}
              onClick={() => void toggleNotifications()}
              className="mt-5 rounded-md border border-edge px-4 py-2 text-sm font-medium text-slate-200 transition-colors hover:bg-edge/60 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {settings?.notificationsEnabled ? "Disable system notifications" : "Enable system notifications"}
            </button>
            <p className="mt-3 text-xs text-slate-400" role="status">
              {settings?.notificationsEnabled ? "System notifications are enabled." : "Notifications are off until you enable them."}
            </p>
          </div>
        </>
      )}

      {saved && <p className="mt-4 text-sm text-emerald-300" role="status">Settings saved.</p>}
      {error && (
        <p className="mt-4 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
          {error}
        </p>
      )}
    </section>
  );
}
