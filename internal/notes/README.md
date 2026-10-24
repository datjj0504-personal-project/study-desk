# internal/notes

Notes domain types and the repository contract live here. The domain package
must not import Wails or SQLite; implementations of `Repository` belong in
`internal/storage`.

`Note.Validate` requires a non-blank title; the body may be empty and the
subject is optional. Repository implementations return `ErrNotFound` when an
operation targets a note that does not exist. `List` returns notes ordered by
most recently updated first.
