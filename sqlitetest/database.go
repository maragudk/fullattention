// Package sqltest provides testing helpers for the sql package.
package sqlitetest

import (
	"testing"

	"maragu.dev/glue/sqlitetest"

	"app/sqlite"
)

// NewDatabase for testing.
func NewDatabase(t *testing.T) *sqlite.Database {
	t.Helper()

	h := sqlitetest.NewHelper(t)
	db := sqlite.NewDatabase(sqlite.NewDatabaseOptions{H: h})
	if err := h.Connect(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := db.H.MigrateUp(t.Context()); err != nil {
		t.Fatal(err)
	}

	return db
}
