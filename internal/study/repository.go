package study

import "context"

type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	CreateCompletedSession(ctx context.Context, session Session, completionKey string) (Session, error)
	GetSession(ctx context.Context, id int64) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
	GetStatistics(ctx context.Context, window StatisticsWindow) (Statistics, error)
}
