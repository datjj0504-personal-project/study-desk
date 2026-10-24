package study

import "context"

type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	GetSession(ctx context.Context, id int64) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
}
