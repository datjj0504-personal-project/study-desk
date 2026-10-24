import { useCallback, useEffect, useState } from "react";
import { Sidebar } from "./components/Sidebar";
import { PomodoroNotificationMonitor } from "./components/PomodoroNotificationMonitor";
import { ComingSoonPage } from "./pages/ComingSoonPage";
import { DashboardPage } from "./pages/DashboardPage";
import { NotesPage } from "./pages/NotesPage";
import { SessionHistoryPage } from "./pages/SessionHistoryPage";
import { StatisticsPage } from "./pages/StatisticsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { TimerPage } from "./pages/TimerPage";
import { timerSettings, type TimerSettings } from "./services/settings";
import { NAV_ITEMS, type PageId } from "./types/navigation";
import { toErrorMessage } from "./utils/errors";

function App() {
  const [page, setPage] = useState<PageId>("dashboard");
  const [settings, setSettings] = useState<TimerSettings | null>(null);
  const [settingsError, setSettingsError] = useState<string | null>(null);
  const [notificationMonitorError, setNotificationMonitorError] = useState<string | null>(null);
  const [notificationDeliveryError, setNotificationDeliveryError] = useState<string | null>(null);
  const [completionNotice, setCompletionNotice] = useState<string | null>(null);
  const current = NAV_ITEMS.find((item) => item.id === page);

  useEffect(() => {
    let mounted = true;
    void timerSettings.get().then((value) => {
      if (mounted) {
        setSettings(value);
        setSettingsError(null);
      }
    }).catch((error: unknown) => {
      if (mounted) setSettingsError(toErrorMessage(error));
    });
    return () => {
      mounted = false;
    };
  }, []);

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
              className="text-slate-400 hover:text-white"
            >
              Dismiss
            </button>
          </div>
        )}
        {(settingsError || notificationDeliveryError || notificationMonitorError) && (
          <p className="mb-5 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
            {settingsError ?? notificationDeliveryError ?? notificationMonitorError}
          </p>
        )}
        {page === "dashboard" ? (
          <DashboardPage />
        ) : page === "notes" ? (
          <NotesPage />
        ) : page === "timer" ? (
          <TimerPage mode="pomodoro" />
        ) : page === "stopwatch" ? (
          <TimerPage mode="stopwatch" />
        ) : page === "settings" ? (
          <SettingsPage settings={settings} onSettingsChange={onSettingsChange} />
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
