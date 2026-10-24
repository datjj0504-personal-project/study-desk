# services/

Thin wrappers around the generated Wails bindings. The only place frontend code should import from `wailsjs/`, so the rest of the UI stays testable. `notes.ts` exposes typed note CRUD and subject operations.
