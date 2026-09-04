package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migration runner for AI Avengers database.
//
// Usage:
//   go run cmd/migrate/main.go up       -- apply all pending migrations
//   go run cmd/migrate/main.go down     -- rollback last migration
//   go run cmd/migrate/main.go down N   -- rollback N migrations
//   go run cmd/migrate/main.go version  -- show current version
//   go run cmd/migrate/main.go force N  -- force set version (use with caution)
//
// WHY golang-migrate:
// Industry standard, supports PostgreSQL, file-based migrations,
// tracks applied migrations in schema_migrations table automatically.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|version|force]")
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	// Migrations are in backend-go/migrations/
	migrationsPath := "file://migrations"
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		migrationsPath = "file://" + envPath
	}

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	command := os.Args[1]

	switch command {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No pending migrations")
				return
			}
			fmt.Fprintf(os.Stderr, "migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Steps(-1); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Nothing to rollback")
				return
			}
			fmt.Fprintf(os.Stderr, "migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration rolled back")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("No migrations applied yet")
				return
			}
			fmt.Fprintf(os.Stderr, "get version failed: %v\n", err)
			os.Exit(1)
		}
		if dirty {
			fmt.Printf("Version: %d (DIRTY - migration failed, fix manually)\n", version)
		} else {
			fmt.Printf("Version: %d\n", version)
		}

	case "force":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "force requires version number: migrate force N")
			os.Exit(1)
		}
		var version int
		if _, err := fmt.Sscanf(os.Args[2], "%d", &version); err != nil {
			fmt.Fprintf(os.Stderr, "invalid version: %s\n", os.Args[2])
			os.Exit(1)
		}
		if err := m.Force(version); err != nil {
			fmt.Fprintf(os.Stderr, "force failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Forced version to %d\n", version)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Println("Usage: migrate [up|down|version|force]")
		os.Exit(1)
	}
}
