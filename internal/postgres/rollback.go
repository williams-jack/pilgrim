package postgres

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/williams-jack/pilgrim/internal/migrations"
	"github.com/williams-jack/pilgrim/internal/parse"
	"github.com/williams-jack/pilgrim/internal/shared"
)

// Arguments for rolling back migrations.
type RollbackArgs struct {
	MigrationName     string
	ConnectionString  string
	DownDir           string
	ConnectionTimeout time.Duration
	MigrationTimeout  time.Duration
}

// Rolls back migrations up to and including the specified migration name.
// If no migration name is specified, rolls back the most recent migration.
func RollbackToMigration(r *RollbackArgs) ([]string, error) {
	migrationPaths, err := migrations.FindMigrations(r.DownDir, shared.DbTypePostgres)
	slices.Reverse(migrationPaths)
	if err != nil {
		return nil, err
	}
	if len(migrationPaths) == 0 {
		return nil, nil
	}
	migrationNames := make([]string, 0)
	for _, path := range migrationPaths {
		name := migrations.MigrationNameFromPath(path, shared.DbTypePostgres)
		if name == "" {
			return nil, fmt.Errorf("could not grab migration name from file: %s", path)
		}
		migrationNames = append(migrationNames, name)
	}
	if r.MigrationName != "" && !slices.Contains(migrationNames, r.MigrationName) {
		return nil, fmt.Errorf("migration name %s not found in down migrations directory", r.MigrationName)
	}
	connTimeout, mTimeout := r.ConnectionTimeout, r.MigrationTimeout
	if connTimeout == 0 {
		connTimeout = shared.DefaultConnectionTimeout
	}
	if mTimeout == 0 {
		mTimeout = shared.DefaultMigrationTimeout
	}
	connCtx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	conn, err := pgx.Connect(connCtx, r.ConnectionString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	unapplied, err := migrationNamesUnapplied(conn, migrationNames, shared.DefaultMiscOpTimeout)
	if err != nil {
		return nil, err
	}

	migrationsRolledBack := make([]string, 0)
	for i, migrationPath := range migrationPaths {
		migrationName := migrationNames[i]
		if unapplied.Contains(migrationName) {
			continue
		}
		f, err := os.Open(migrationPath)
		if err != nil {
			return migrationsRolledBack, err
		}
		defer f.Close()
		sqlWDirectives, err := parse.ParseSQLFile(f)
		if err != nil {
			return migrationsRolledBack, err
		}
		err = rollbackMigration(conn, migrationName, sqlWDirectives, mTimeout)
		if err != nil {
			return migrationsRolledBack, err
		}
		migrationsRolledBack = append(migrationsRolledBack, migrationName)
		if migrationName == r.MigrationName || r.MigrationName == "" {
			break
		}
	}
	return migrationsRolledBack, nil
}

// PgRollbackMigration rolls back a previously applied migration by executing
// the provided SQL and removing the migration record from the history table.
// Returns true if the migration was previously applied.
func rollbackMigration(conn *pgx.Conn, migrationName string, sqlWDirectives *parse.SQLWithDirectives, queryTimeout time.Duration) error {
	ctx, queryCancel := context.WithTimeout(context.Background(), queryTimeout)
	defer queryCancel()
	_, err := conn.Exec(ctx, "BEGIN;")
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, sqlWDirectives.Sql)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
		DELETE FROM pilgrim_migration_history
		WHERE migration_name = $1;
	`, migrationName)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, "COMMIT;")
	return err
}
