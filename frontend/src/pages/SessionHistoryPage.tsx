import { useCallback, useEffect, useMemo, useState } from "react";
import { listSubjects, type Subject } from "../services/notes";
import { listStudySessions, type SessionKind, type StudySession } from "../services/sessions";
import { toErrorMessage } from "../utils/errors";

const kindLabels: Record<SessionKind, string> = {
  pomodoro: "Pomodoro",
  stopwatch: "Stopwatch",
  manual: "Manual",
};

function localDateKey(value: string): string {
  const date = new Date(value);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatDuration(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) {
    return seconds > 0
      ? `${hours} hr ${minutes} min ${seconds} sec`
      : `${hours} hr ${minutes} min`;
  }
  if (minutes > 0) return seconds > 0 ? `${minutes} min ${seconds} sec` : `${minutes} min`;
  return `${seconds} sec`;
}

function formatDate(dateKey: string): string {
  const [year, month, day] = dateKey.split("-").map(Number);
  return new Date(year, month - 1, day).toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function formatTime(value: string): string {
  return new Date(value).toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" });
}

function parseKindFilter(value: string): SessionKind | "all" {
  if (value === "pomodoro" || value === "stopwatch" || value === "manual") return value;
  return "all";
}

interface SessionGroup {
  date: string;
  sessions: StudySession[];
}

export function SessionHistoryPage() {
  const [sessions, setSessions] = useState<StudySession[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [kindFilter, setKindFilter] = useState<SessionKind | "all">("all");
  const [subjectFilter, setSubjectFilter] = useState<number | "all">("all");

  const loadHistory = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [loadedSessions, loadedSubjects] = await Promise.all([
        listStudySessions(),
        listSubjects(),
      ]);
      setSessions(loadedSessions);
      setSubjects(loadedSubjects);
    } catch (loadError) {
      setError(toErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadHistory();
  }, [loadHistory]);

  const subjectNames = useMemo(
    () => new Map(subjects.map((subject) => [subject.id, subject.name])),
    [subjects],
  );

  const groups = useMemo(() => {
    const filtered = sessions.filter((session) => {
      const date = localDateKey(session.startedAt);
      return (
        (dateFrom === "" || date >= dateFrom) &&
        (dateTo === "" || date <= dateTo) &&
        (kindFilter === "all" || session.kind === kindFilter) &&
        (subjectFilter === "all" || session.subjectId === subjectFilter)
      );
    });
    const grouped = new Map<string, StudySession[]>();
    for (const session of filtered) {
      const date = localDateKey(session.startedAt);
      const items = grouped.get(date) ?? [];
      items.push(session);
      grouped.set(date, items);
    }
    return Array.from(grouped, ([date, items]): SessionGroup => ({ date, sessions: items }));
  }, [dateFrom, dateTo, kindFilter, sessions, subjectFilter]);

  const hasFilters = dateFrom !== "" || dateTo !== "" || kindFilter !== "all" || subjectFilter !== "all";

  function clearFilters() {
    setDateFrom("");
    setDateTo("");
    setKindFilter("all");
    setSubjectFilter("all");
  }

  return (
    <section className="mx-auto max-w-5xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Session History</h1>
          <p className="mt-1 text-slate-400">Browse recorded study sessions by date.</p>
        </div>
        <button
          type="button"
          onClick={() => void loadHistory()}
          disabled={loading}
          className="rounded-md border border-edge px-3 py-2 text-sm hover:bg-edge disabled:opacity-50"
        >
          Refresh
        </button>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-3 rounded-xl border border-edge bg-panel p-4 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label htmlFor="session-date-from" className="mb-1 block text-xs font-medium text-slate-400">From</label>
          <input
            id="session-date-from"
            type="date"
            value={dateFrom}
            max={dateTo || undefined}
            onChange={(event) => setDateFrom(event.target.value)}
            className="w-full rounded-md border border-edge bg-surface px-2 py-2 text-sm"
          />
        </div>
        <div>
          <label htmlFor="session-date-to" className="mb-1 block text-xs font-medium text-slate-400">To</label>
          <input
            id="session-date-to"
            type="date"
            value={dateTo}
            min={dateFrom || undefined}
            onChange={(event) => setDateTo(event.target.value)}
            className="w-full rounded-md border border-edge bg-surface px-2 py-2 text-sm"
          />
        </div>
        <div>
          <label htmlFor="session-kind-filter" className="mb-1 block text-xs font-medium text-slate-400">Session type</label>
          <select
            id="session-kind-filter"
            value={kindFilter}
            onChange={(event) => setKindFilter(parseKindFilter(event.target.value))}
            className="w-full rounded-md border border-edge bg-surface px-2 py-2 text-sm"
          >
            <option value="all">All types</option>
            <option value="pomodoro">Pomodoro</option>
            <option value="stopwatch">Stopwatch</option>
            <option value="manual">Manual</option>
          </select>
        </div>
        <div>
          <label htmlFor="session-subject-filter" className="mb-1 block text-xs font-medium text-slate-400">Subject</label>
          <select
            id="session-subject-filter"
            value={subjectFilter}
            onChange={(event) => setSubjectFilter(event.target.value === "all" ? "all" : Number(event.target.value))}
            className="w-full rounded-md border border-edge bg-surface px-2 py-2 text-sm"
          >
            <option value="all">All subjects</option>
            {subjects.map((subject) => (
              <option key={subject.id} value={subject.id}>{subject.name}</option>
            ))}
          </select>
        </div>
        {hasFilters && (
          <button
            type="button"
            onClick={clearFilters}
            className="text-left text-sm text-indigo-300 hover:text-indigo-200 sm:col-span-2 lg:col-span-4"
          >
            Clear filters
          </button>
        )}
      </div>

      {loading && sessions.length === 0 ? (
        <p className="mt-6 text-slate-400" role="status">Loading session history…</p>
      ) : null}
      {error && (
        <div className="mt-6 flex items-center justify-between gap-4 rounded-md border border-red-500/40 bg-red-950/40 px-4 py-3 text-sm text-red-200" role="alert">
          <span>{error}</span>
          {sessions.length > 0 && <span className="shrink-0">Showing previously loaded sessions.</span>}
        </div>
      )}
      {!loading && !error && sessions.length === 0 && (
        <p className="mt-6 rounded-xl border border-edge bg-panel p-5 text-slate-400">
          No study sessions have been recorded yet.
        </p>
      )}
      {!loading && !error && sessions.length > 0 && groups.length === 0 && (
        <p className="mt-6 rounded-xl border border-edge bg-panel p-5 text-slate-400" role="status">
          No sessions match these filters.
        </p>
      )}

      <div className="mt-6 space-y-6">
        {groups.map((group) => (
          <section key={group.date} aria-label={formatDate(group.date)}>
            <h2 className="mb-3 text-lg font-semibold">{formatDate(group.date)}</h2>
            <ul className="space-y-3">
              {group.sessions.map((session) => (
                <li key={session.id} className="rounded-xl border border-edge bg-panel p-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <p className="font-medium">
                        {kindLabels[session.kind]}
                        {session.subjectId !== undefined && (
                          <span className="ml-2 rounded bg-edge px-2 py-0.5 text-xs text-slate-300">
                            {subjectNames.get(session.subjectId) ?? "Deleted subject"}
                          </span>
                        )}
                      </p>
                      <p className="mt-1 text-sm text-slate-400">
                        {formatTime(session.startedAt)} – {formatTime(session.endedAt)}
                      </p>
                    </div>
                    <p className="font-medium tabular-nums">{formatDuration(session.durationSeconds)}</p>
                  </div>
                  {session.note.trim() && (
                    <p className="mt-3 whitespace-pre-wrap border-t border-edge pt-3 text-sm text-slate-300">
                      {session.note}
                    </p>
                  )}
                </li>
              ))}
            </ul>
          </section>
        ))}
      </div>
    </section>
  );
}
