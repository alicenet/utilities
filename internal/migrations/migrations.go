package migrations

import (
	"embed"
	"fmt"
	"testing"

	"cloud.google.com/go/spanner/spannertest"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const (
	emulatorAddr        = "localhost:0"
	emulatorDatabase    = "projects/ignored/instances/ignored/databases/ignored"
	emulatorEnvironment = "SPANNER_EMULATOR_HOST"
	urlFormat           = "spanner://%s?x-clean-statements=true"
)

//go:embed *.sql
var migrationFiles embed.FS

// SetupEmulator for Spanner in a testing environment.
func SetupEmulator(t *testing.T) {
	t.Helper()

	srv, err := spannertest.NewServer(emulatorAddr)
	if err != nil {
		t.Fatalf("emulator: %v", err)
	}

	t.Cleanup(func() {
		srv.Close()
	})

	t.Setenv(emulatorEnvironment, srv.Addr)
	srv.SetLogger(t.Logf)
}

// RunTestMigrations for Spanner in a testing environment.
func RunTestMigrations(t *testing.T) {
	t.Helper()

	driver, err := GetMigrations()
	if err != nil {
		t.Fatal("embedded migrations:", err)
	}

	url := fmt.Sprintf(urlFormat, emulatorDatabase)

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, url)
	if err != nil {
		t.Fatal("migrations:", err)
	}

	if err := migrations.Up(); err != nil {
		t.Fatal("up:", err)
	}

	t.Cleanup(func() {
		if err := migrations.Drop(); err != nil {
			t.Error("drop:", err)
		}
	})
}

// Run migrations for a given Spanner database.
func RunMigrations(database string) error {
	driver, err := GetMigrations()
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	url := fmt.Sprintf(urlFormat, database)

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, url)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	if err := migrations.Up(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// GetMigrations from SQL files.
func GetMigrations() (source.Driver, error) {
	d, err := iofs.New(migrationFiles, ".")
	if err != nil {
		return nil, fmt.Errorf("getting migrations: %w", err)
	}

	return d, nil
}
