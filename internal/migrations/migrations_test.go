package migrations

import (
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/spanner"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//nolint:paralleltest // t.Parallel not supported with t.Setenv
func TestMigrations(t *testing.T) {
	SetupEmulator(t)

	driver, err := GetMigrations()
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf(urlFormat, emulatorDatabase)

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, url)
	if err != nil {
		t.Fatal("migrations:", err)
	}

	// This approach will not catch all potential migration issues as it
	// does not validate step by step, but it's good enough to start.
	if err := migrations.Up(); err != nil {
		t.Fatal("up:", err)
	}

	if err := migrations.Down(); err != nil {
		t.Fatal("down:", err)
	}

	if err := migrations.Up(); err != nil {
		t.Fatal("up again:", err)
	}

	if err := migrations.Drop(); err != nil {
		t.Fatal("drop:", err)
	}
}
