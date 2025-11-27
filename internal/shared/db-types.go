package shared

import "fmt"

type DbType string

// Database types supported by Pilgrim.
const (
	DbTypePostgres DbType = "postgres"
)

func UnsupportedDbTypeError(unsupportedDbType DbType) error {
	return fmt.Errorf("unsupported database type: %s", unsupportedDbType)
}

// FileExtension returns the standard file extension for migration files
// associated with the database type.
func (dbt DbType) FileExtension() string {
	switch dbt {
	case DbTypePostgres:
		return "sql"
	default:
		return ""
	}
}

// Returns a list of all database types supported by Pilgrim.
func SupportedDbTypes() []DbType {
	return []DbType{
		DbTypePostgres,
	}
}
