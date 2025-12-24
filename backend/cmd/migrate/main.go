package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Migration represents a single migration
type Migration struct {
	Version     string
	Description string
	Up          func(ctx context.Context, tx pgx.Tx) error
	Down        func(ctx context.Context, tx pgx.Tx) error
}

// Migrator handles database migrations
type Migrator struct {
	db         *pgxpool.Pool
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *pgxpool.Pool) *Migrator {
	return &Migrator{
		db: db,
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(version, description string, up, down func(ctx context.Context, tx pgx.Tx) error) {
	m.migrations = append(m.migrations, Migration{
		Version:     version,
		Description: description,
		Up:          up,
		Down:        down,
	})
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	_, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := m.db.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s: %s", migration.Version, migration.Description)

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if err := migration.Up(ctx, tx); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		// Record the migration
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s applied successfully", migration.Version)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations in reverse order
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if !applied[migration.Version] {
			continue
		}

		log.Printf("Rolling back migration %s: %s", migration.Version, migration.Description)

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if err := migration.Down(ctx, tx); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("rollback migration %s failed: %w", migration.Version, err)
		}

		// Remove the migration record
		_, err = tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to remove migration record %s: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s rolled back successfully", migration.Version)
		return nil
	}

	log.Println("No migrations to rollback")
	return nil
}

// Status shows the migration status
func (m *Migrator) Status(ctx context.Context) error {
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			fmt.Printf("%s - Applied\n", migration.Version)
		} else {
			fmt.Printf("%s - Pending\n", migration.Version)
		}
	}
	fmt.Println()

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate <command> [database-url]")
	}

	command := os.Args[1]
	// Load .env so DATABASE_URL from backend/.env is available
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if len(os.Args) > 2 {
		dbURL = os.Args[2]
	}

	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable or argument is required")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	migrator := NewMigrator(pool)

	// Register all migrations
	registerMigrations(migrator)

	switch command {
	case "up":
		if err := migrator.Up(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("All migrations completed successfully!")

	case "down":
		if err := migrator.Down(ctx); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully!")

	case "status":
		if err := migrator.Status(ctx); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}

	default:
		log.Fatalf("Unknown command: %s. Use 'up', 'down', or 'status'", command)
	}
}
