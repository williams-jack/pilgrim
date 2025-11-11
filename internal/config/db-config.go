package config

// DBConfig is an interface that defines methods for database configuration.
type DbConfig interface {
	// ConnectionString returns the database connection string. This is often
	// either provided directly or constructed from individual configuration
	// components.
	ConnectionString() string
}
