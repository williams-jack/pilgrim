package postgres

import (
	"fmt"
	"strings"
)

// PgConfig holds the configuration details for connecting to a PostgreSQL database.
type PgConfig struct {
	// Databse hostname or IP address
	Host string `json:"host"`
	// Database port number
	Port int `json:"port"`
	// Username for authentication
	User string `json:"user"`
	// Password for authentication
	Password string `json:"password"`
	// Name of the database to connect to
	DbName string `json:"dbName"`
	// Additional connection parameters
	Params map[string]string `json:"params"`
}

// ConnectionString constructs the PostgreSQL connection string from the PgConfig fields.
func (pg PgConfig) ConnectionString() string {
	var connStrBuilder strings.Builder
	connStrBuilder.WriteString("postgres://")
	if pg.User != "" {
		connStrBuilder.WriteString(pg.User)
		if pg.Password != "" {
			connStrBuilder.WriteString(":" + pg.Password)
		}
		connStrBuilder.WriteString("@")
	}
	if pg.Host != "" {
		connStrBuilder.WriteString(pg.Host)
	}
	if pg.Port != 0 {
		connStrBuilder.WriteString(fmt.Sprintf(":%d", pg.Port))
	}
	if pg.DbName != "" {
		connStrBuilder.WriteString("/" + pg.DbName)
	}
	if len(pg.Params) > 0 {
		var params []string
		for key, value := range pg.Params {
			params = append(params, fmt.Sprintf("%s=%s", key, value))
		}
		connStrBuilder.WriteString("?" + strings.Join(params, "&"))
	}
	return connStrBuilder.String()
}
