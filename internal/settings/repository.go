package settings

import "context"

type Repository interface {
	LoadTimerSettings(ctx context.Context) (TimerSettings, error)
	SaveTimerSettings(ctx context.Context, settings TimerSettings) error
	LoadTheme(ctx context.Context) (Theme, error)
	SaveTheme(ctx context.Context, theme Theme) error
}
