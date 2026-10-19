package storage

import (
	"context"
	"testing"
)

func TestOpenSQLite(t *testing.T) {
	db, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping SQLite database: %v", err)
	}

	var version string
	if err := db.QueryRowContext(context.Background(), "SELECT sqlite_version()").Scan(&version); err != nil {
		t.Fatalf("query SQLite version: %v", err)
	}
	if version == "" {
		t.Fatal("SQLite returned an empty version")
	}
}
