# internal/settings

Owns user timer and appearance preferences, defaults, and validation. Timer
durations are converted to timer values here. Settings persistence is provided
by `internal/storage`; this package must not import Wails.

Pomodoro durations are configured in whole minutes: focus is limited to 1–180,
breaks to 1–60, and the long-break interval to 1–12 focus sessions.
The theme preference may be `system`, `light`, or `dark`; `dark` is the default
to preserve the application's existing appearance.
