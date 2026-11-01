# internal/notes

Notes domain types and the repository contract live here. The domain package
must not import Wails or SQLite; implementations of `Repository` belong in
`internal/storage`.

`Note.Validate` requires a non-blank title; the body may be empty and the
subject is optional. `Create` assigns the ID and UTC timestamps. `Update`
preserves creation time and refreshes update time. Repository implementations
return `ErrNotFound` when an operation targets a note that does not exist.
`List` returns notes ordered by most recently updated first.

Subjects group notes. Subject names must not be blank; repositories create
subjects and list them alphabetically. Notes may reference one subject or none.
