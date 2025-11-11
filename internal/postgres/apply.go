package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/williams-jack/pilgrim/internal/migrations"
	"github.com/williams-jack/pilgrim/internal/parse"
	"github.com/williams-jack/pilgrim/internal/shared"
)

type ApplyMigrationsArgs struct {
	ConnectionString         string
	UpDir                    string
	MigrationName            string
	ConnectionTimeout        time.Duration
	MigrationTimeout         time.Duration
	InternalOperationTimeout time.Duration
}

// Applies all migrations up to and including the specified migration name.
// If migrationName is empty, applies all pending migrations.
// Returns a list of all the migration files successfully applied, and an
// error if one occurred.
func ApplyMigrations(args *ApplyMigrationsArgs) ([]string, error) {
	connTimeout := args.ConnectionTimeout
	if connTimeout == 0 {
		connTimeout = shared.DefaultConnectionTimeout
	}
	mTimeout := args.MigrationTimeout
	if mTimeout == 0 {
		mTimeout = shared.DefaultMigrationTimeout
	}
	internalOpTimeout := args.InternalOperationTimeout
	if internalOpTimeout == 0 {
		internalOpTimeout = shared.DefaultMiscOpTimeout
	}
	migrationPaths, err := migrations.FindMigrations(
		args.UpDir,
		shared.DbTypePostgres,
	)
	migrationNames := make([]string, 0, len(migrationPaths))
	for _, p := range migrationPaths {
		name := migrations.MigrationNameFromPath(p, shared.DbTypePostgres)
		if name == "" {
			return nil, fmt.Errorf(
				"cannot get migration name from file %v",
				p,
			)
		}
		migrationNames = append(migrationNames, name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	conn, err := pgx.Connect(ctx, args.ConnectionString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())
	unappliedMigrationNamesSet, err := migrationNamesUnapplied(conn, migrationNames, internalOpTimeout)
	if err != nil {
		return nil, err
	}

	appliedMigrationFiles := make([]string, 0)
	for i, mPath := range migrationPaths {
		mName := migrationNames[i]
		if !unappliedMigrationNamesSet.Contains(mName) {
			continue
		}
		f, err := os.Open(mPath)
		if err != nil {
			return appliedMigrationFiles, err
		}
		sqlWDirectives, err := parse.ParseSQLFile(f)
		if err != nil {
			return appliedMigrationFiles, err
		}
		err = applyMigration(conn, mName, sqlWDirectives, mTimeout)
		if err != nil {
			return appliedMigrationFiles, err
		}
		appliedMigrationFiles = append(appliedMigrationFiles, mPath)
	}
	return appliedMigrationFiles, nil
}

// Applies a migration by executing the provided SQL (and Pilgrim directives)
// and recording the migration in the history table. Returns true if the migratiion already existed.
func applyMigration(conn *pgx.Conn, migrationName string, sqlWDirectives *parse.SQLWithDirectives, migrationTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)

	defer cancel()

	_, err := conn.Exec(ctx, "BEGIN;")
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, sqlWDirectives.Sql)
	if err != nil {
		return err
	}
	err = pgRecordMigration(conn, ctx, migrationName, sqlWDirectives.Directives.Description)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, "COMMIT;")
	return err
}

// Helper function to record a migration in the migration history table.
func pgRecordMigration(conn *pgx.Conn, ctx context.Context, migrationName, migrationDescription string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err := conn.Exec(ctx, `
		INSERT INTO pilgrim_migration_history (migration_name, migration_description)
		VALUES ($1, $2);
	`, migrationName, migrationDescription)
	return err
}
