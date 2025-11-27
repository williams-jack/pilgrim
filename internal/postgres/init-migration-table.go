package postgres

import (
	"context"
	pgx "github.com/jackc/pgx/v5"
	"github.com/williams-jack/pilgrim/internal/shared"
)

// PgInitMigrationTable initializes the migration history table in the PostgreSQL database
// specified by the connection string.
func PgInitMigrationTable(connectionString string) error {
	// TODO: Do we need to handle SSL mode here?
	// TODO: Do we care about timeout for initial connection?
	ctx, cancel := context.WithTimeout(context.Background(), shared.DefaultConnectionTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return initializePGDatabase(conn)
}

// Helper function to initialize the migration history table if it doesn't exist.
func initializePGDatabase(conn *pgx.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), shared.DefaultConnectionTimeout)
	defer cancel()
	_, err := conn.Exec(ctx, `
	BEGIN;
	CREATE TABLE IF NOT EXISTS pilgrim_migration_history (
		id SERIAL PRIMARY KEY,
		migration_name TEXT NOT NULL,
		migration_description TEXT DEFAULT '',
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_migration_name UNIQUE (migration_name)
	);
	COMMIT;
	`)
	return err
}
