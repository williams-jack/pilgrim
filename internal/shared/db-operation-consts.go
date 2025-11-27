package shared

import "time"

const (
	// Default time for Pilgrim to wait in establishing a database connection before giving up.
	DefaultConnectionTimeout time.Duration = 5 * time.Second
	// Default time (total) Pilgrim will wait while attempting to run a migration before cancelling execution.
	DefaultMigrationTimeout  time.Duration = 30 * time.Second
	// Default timeout Pilgrim will wait on internal operations (e.g., grabbing all applied migrations) before
	// cancelling their execution.
	DefaultMiscOpTimeout time.Duration = 5 * time.Second
)
