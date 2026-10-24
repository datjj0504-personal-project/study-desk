# internal/settings

Owns user timer preferences, defaults, validation, and conversion to timer
durations. Settings persistence is provided by `internal/storage`; this package
must not import Wails.

Pomodoro durations are configured in whole minutes: focus is limited to 1–180,
breaks to 1–60, and the long-break interval to 1–12 focus sessions.
