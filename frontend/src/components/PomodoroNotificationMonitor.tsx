import { useEffect, useRef } from "react";
import { pomodoroTimer } from "../services/timers";

interface PomodoroNotificationMonitorProps {
  onCompletion: (phase: string) => void;
  onError: (message: string | null) => void;
}

export function PomodoroNotificationMonitor({
  onCompletion,
  onError,
}: PomodoroNotificationMonitorProps) {
  const completionHandler = useRef(onCompletion);
  completionHandler.current = onCompletion;

  useEffect(() => {
    let mounted = true;
    let timeout: ReturnType<typeof setTimeout>;
    let previousStatus: string | null = null;

    const poll = async () => {
      try {
        const state = await pomodoroTimer.getState();
        if (!mounted) return;
        if (previousStatus !== null && previousStatus !== "completed" && state.status === "completed") {
          completionHandler.current(state.phase);
        }
        previousStatus = state.status;
        onError(null);
      } catch (pollError) {
        if (mounted) onError(pollError instanceof Error ? pollError.message : String(pollError));
      } finally {
        if (mounted) timeout = setTimeout(poll, 1000);
      }
    };

    void poll();
    return () => {
      mounted = false;
      clearTimeout(timeout);
    };
  }, [onError]);

  return null;
}
