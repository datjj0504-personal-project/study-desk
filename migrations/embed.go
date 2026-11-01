package migrations

import "embed"

// Files contains the versioned SQL migrations bundled with the application.
//
//go:embed *.sql
var Files embed.FS
