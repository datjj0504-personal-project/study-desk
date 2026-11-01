package notes

import "context"

// Repository persists notes. Implementations return ErrNotFound from
// GetByID, Update, and Delete when the requested note does not exist. List
// orders notes by UpdatedAt descending.
type Repository interface {
	Create(context.Context, Note) (Note, error)
	GetByID(context.Context, int64) (Note, error)
	Update(context.Context, Note) error
	Delete(context.Context, int64) error
	List(context.Context) ([]Note, error)
	CreateSubject(context.Context, Subject) (Subject, error)
	ListSubjects(context.Context) ([]Subject, error)
}
