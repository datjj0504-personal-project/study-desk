import { useCallback, useEffect, useRef, useState } from "react";
import { Sidebar } from "./components/Sidebar";
import { PomodoroNotificationMonitor } from "./components/PomodoroNotificationMonitor";
import { ComingSoonPage } from "./pages/ComingSoonPage";
import { DashboardPage } from "./pages/DashboardPage";
import { NotesPage } from "./pages/NotesPage";
import { SessionHistoryPage } from "./pages/SessionHistoryPage";
import { StatisticsPage } from "./pages/StatisticsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { TimerPage } from "./pages/TimerPage";
import {
  themeSettings,
  timerSettings,
  type Theme,
  type TimerSettings,
} from "./services/settings";
import { NAV_ITEMS, type PageId } from "./types/navigation";
import { toErrorMessage } from "./utils/errors";

function App() {
  const [page, setPage] = useState<PageId>("dashboard");
  const [settings, setSettings] = useState<TimerSettings | null>(null);
  const [theme, setTheme] = useState<Theme | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [themeLoading, setThemeLoading] = useState(true);
  const [settingsError, setSettingsError] = useState<string | null>(null);
  const [themeError, setThemeError] = useState<string | null>(null);
  const [notificationMonitorError, setNotificationMonitorError] = useState<string | null>(null);
  const [notificationDeliveryError, setNotificationDeliveryError] = useState<string | null>(null);
  const [completionNotice, setCompletionNotice] = useState<string | null>(null);
  const mountedRef = useRef(false);
  const current = NAV_ITEMS.find((item) => item.id === page);

  const loadTimerSettings = useCallback(async () => {
    setSettingsLoading(true);
    setSettingsError(null);
    try {
      const value = await timerSettings.get();
      if (mountedRef.current) setSettings(value);
    } catch (error) {
      if (mountedRef.current) setSettingsError(toErrorMessage(error));
    } finally {
      if (mountedRef.current) setSettingsLoading(false);
    }
  }, []);

  const loadTheme = useCallback(async () => {
    setThemeLoading(true);
    setThemeError(null);
    try {
      const value = await themeSettings.get();
      if (mountedRef.current) setTheme(value);
    } catch (error) {
      if (mountedRef.current) setThemeError(toErrorMessage(error));
    } finally {
      if (mountedRef.current) setThemeLoading(false);
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    void loadTimerSettings();
    void loadTheme();
    return () => {
      mountedRef.current = false;
    };
  }, [loadTheme, loadTimerSettings]);

  useEffect(() => {
    if (theme) document.documentElement.dataset.theme = theme;
  }, [theme]);

  const onPhaseCompletion = useCallback((phase: string) => {
    const label = phase === "focus"
      ? "Focus session"
      : phase === "short_break"
        ? "Short break"
        : phase === "long_break"
          ? "Long break"
          : "Pomodoro phase";
    setCompletionNotice(`${label} complete.`);
    if (!settings?.notificationsEnabled) return;
    if (typeof Notification === "undefined") {
      setNotificationDeliveryError("System notifications are not available in this environment.");
      return;
    }
    if (Notification.permission !== "granted") {
      setNotificationDeliveryError("System notification permission is no longer granted. Re-enable it in Settings.");
      return;
    }
    try {
      new Notification("StudyDesk", { body: `${label} complete.` });
      setNotificationDeliveryError(null);
    } catch (error) {
      setNotificationDeliveryError(toErrorMessage(error));
    }
  }, [settings?.notificationsEnabled]);

  const onMonitorError = useCallback((message: string | null) => {
    setNotificationMonitorError(message);
  }, []);

  const onSettingsChange = useCallback((value: TimerSettings) => {
    setSettings(value);
    if (!value.notificationsEnabled) setNotificationDeliveryError(null);
  }, []);

  const onThemeChange = useCallback((value: Theme) => {
    setTheme(value);
    setThemeError(null);
  }, []);

  return (
    <div className="flex h-full">
      <PomodoroNotificationMonitor
        onCompletion={onPhaseCompletion}
        onError={onMonitorError}
      />
      <Sidebar active={page} onSelect={setPage} />
      <main className="flex-1 overflow-y-auto p-8">
        {completionNotice && (
          <div className="mb-5 flex items-center justify-between gap-4 rounded-md border border-accent/50 bg-panel px-4 py-3 text-sm" role="status">
            <span>{completionNotice}</span>
            <button
              type="button"
              aria-label="Dismiss timer notification"
              onClick={() => setCompletionNotice(null)}
              className="text-slate-400 hover:text-slate-900"
            >
              Dismiss
            </button>
          </div>
        )}
        <div className="mb-5 space-y-2">
          {settingsError && (
            <div className="flex items-center justify-between gap-4 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
              <span>Could not load timer settings: {settingsError}</span>
              <button
                type="button"
                disabled={settingsLoading}
                onClick={() => void loadTimerSettings()}
                className="shrink-0 rounded border border-red-200/40 px-3 py-1.5 hover:bg-red-900/40 disabled:opacity-50"
              >
                {settingsLoading ? "Retrying…" : "Retry"}
              </button>
            </div>
          )}
          {themeError && (
            <div className="flex items-center justify-between gap-4 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
              <span>Could not load theme preference: {themeError}</span>
              <button
                type="button"
                disabled={themeLoading}
                onClick={() => void loadTheme()}
                className="shrink-0 rounded border border-red-200/40 px-3 py-1.5 hover:bg-red-900/40 disabled:opacity-50"
              >
                {themeLoading ? "Retrying…" : "Retry"}
              </button>
            </div>
          )}
          {notificationDeliveryError && (
            <p className="rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
              {notificationDeliveryError}
            </p>
          )}
          {notificationMonitorError && (
            <p className="rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
              Timer status unavailable: {notificationMonitorError}
            </p>
          )}
        </div>
        {page === "dashboard" ? (
          <DashboardPage />
        ) : page === "notes" ? (
          <NotesPage />
        ) : page === "timer" ? (
          <TimerPage mode="pomodoro" />
        ) : page === "stopwatch" ? (
          <TimerPage mode="stopwatch" />
        ) : page === "settings" ? (
          <SettingsPage
            settings={settings}
            onSettingsChange={onSettingsChange}
            theme={theme}
            onThemeChange={onThemeChange}
          />
        ) : page === "sessions" ? (
          <SessionHistoryPage />
        ) : page === "statistics" ? (
          <StatisticsPage />
        ) : (
          <ComingSoonPage title={current?.label ?? "Page"} />
        )}
      </main>
    </div>
  );
}

export default App;
