package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/williams-jack/pilgrim/internal/shared"
)


func migrationNamesUnapplied(conn *pgx.Conn, migrationNames []string, t time.Duration) (*shared.Set[string], error) {
	appliedMigrations := shared.NewSet(migrationNames...)
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	q := "SELECT migration_name " +
		"FROM pilgrim_migration_history;"
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrationName string
	_, err = pgx.ForEachRow(rows, []any{&migrationName}, func() error {
		appliedMigrations.Remove(migrationName)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return appliedMigrations, nil
}
