# services/

Thin wrappers around the generated Wails bindings. The only place frontend code should import from `wailsjs/`, so the rest of the UI stays testable. `notes.ts` exposes typed note CRUD and subject operations; `timers.ts` exposes typed Pomodoro and stopwatch snapshots and controls; `settings.ts` exposes persisted Pomodoro and appearance preferences; `statistics.ts` exposes persisted study-duration aggregates; `sessions.ts` exposes persisted study-session history.
