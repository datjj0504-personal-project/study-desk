import { useCallback, useEffect, useState } from "react";
import { getStudyStatistics, type StudyStatistics } from "../services/statistics";
import { toErrorMessage } from "../utils/errors";

function formatDuration(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) {
    return seconds > 0
      ? `${hours} hr ${minutes} min ${seconds} sec`
      : `${hours} hr ${minutes} min`;
  }
  if (minutes > 0) {
    return seconds > 0 ? `${minutes} min ${seconds} sec` : `${minutes} min`;
  }
  return `${seconds} sec`;
}

const totalCards = [
  { key: "todaySeconds", label: "Today" },
  { key: "weekSeconds", label: "This week" },
  { key: "allTimeSeconds", label: "All time" },
] as const;

export function StatisticsPage() {
  const [statistics, setStatistics] = useState<StudyStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadStatistics = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setStatistics(await getStudyStatistics());
    } catch (loadError) {
      setError(toErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadStatistics();
  }, [loadStatistics]);

  return (
    <section className="mx-auto max-w-4xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Study Statistics</h1>
          <p className="mt-1 text-slate-400">
            Study duration is counted by session start time in your local timezone. Weeks start on Monday.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void loadStatistics()}
          disabled={loading}
          className="rounded-md border border-edge px-3 py-2 text-sm hover:bg-edge disabled:opacity-50"
        >
          Refresh
        </button>
      </div>

      {loading && !statistics ? (
        <p className="mt-6 text-slate-400" role="status">Loading statistics…</p>
      ) : error ? (
        <div className="mt-6 rounded-xl border border-red-500/40 bg-red-950/40 p-4 text-red-200" role="alert">
          <p>{error}</p>
          {statistics && <p className="mt-2 text-sm">Previously loaded totals are still shown below.</p>}
        </div>
      ) : null}

      {statistics && (
        <>
          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
            {totalCards.map(({ key, label }) => (
              <article key={key} className="rounded-xl border border-edge bg-panel p-5">
                <h2 className="text-sm font-medium text-slate-400">{label}</h2>
                <p className="mt-2 text-2xl font-semibold">{formatDuration(statistics[key])}</p>
              </article>
            ))}
          </div>

          <section className="mt-8 rounded-xl border border-edge bg-panel p-5">
            <h2 className="text-lg font-semibold">Study time by subject</h2>
            {statistics.bySubject.length === 0 ? (
              <p className="mt-4 text-sm text-slate-400">No completed study sessions yet.</p>
            ) : (
              <ul className="mt-4 divide-y divide-edge">
                {statistics.bySubject.map((subject) => (
                  <li
                    key={subject.subjectId ?? "unassigned"}
                    className="flex items-center justify-between gap-4 py-3 first:pt-0 last:pb-0"
                  >
                    <span>{subject.subjectName}</span>
                    <span className="font-medium tabular-nums">{formatDuration(subject.durationSeconds)}</span>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      )}
    </section>
  );
}
