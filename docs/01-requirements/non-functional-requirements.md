# Non-Functional Requirements

- NFR-1 **Offline-first**: all features work without network access; no telemetry or cloud calls.
- NFR-2 **Responsiveness**: UI interactions respond in under 100 ms; startup under 3 seconds on typical hardware; timers must not block the UI.
- NFR-3 **Data integrity**: data stored in a local SQLite file; writes are transactional; schema changes go through versioned [migrations](../../migrations/README.md); a running session must not be lost on normal app close.
- NFR-4 **Timer accuracy**: elapsed time is computed from timestamps, not by counting ticks, so it stays correct under throttling or sleep.
- NFR-5 **Maintainability**: follow [architecture](../02-design/architecture.md) boundaries and [coding conventions](../03-development/coding-conventions.md); business logic is unit-testable without Wails.
- NFR-6 **Windows compatibility**: supports Windows 10/11 (64-bit) with the WebView2 runtime.
- NFR-7 **Privacy**: user data stays on the device in the user's profile directory.
