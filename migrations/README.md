# migrations/

Versioned SQL migrations for SQLite, named `NNNN_description.sql` (e.g.
`0001_init.sql`). The SQL files are embedded into the application and applied in
order at startup by `internal/storage`. Never edit a migration once released;
add a new one. The schema is described in [database design](../docs/02-design/database-design.md).
