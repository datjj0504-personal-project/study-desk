package timer

import "context"

type StateRepository interface {
	LoadTimerState(ctx context.Context, mode Mode) (Snapshot, bool, error)
	SaveTimerState(ctx context.Context, snapshot Snapshot) error
}
