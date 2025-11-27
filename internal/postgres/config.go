package postgres

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/williams-jack/pilgrim/internal/shared"
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

type pgConfigRaw struct {
	Host     string            `json:"host"`
	Port     any               `json:"port"`
	User     string            `json:"user"`
	Password string            `json:"password"`
	DbName   string            `json:"dbName"`
	Params   map[string]string `json:"params"`
}

// ConnectionString constructs the PostgreSQL connection string from the PgConfig fields.
func (pg *PgConfig) ConnectionString() string {
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

// UnmarshalJSON custom unmarshals the PgConfig from JSON, replacing environment variable
// placeholders with their actual values.
func (pg *PgConfig) UnmarshalJSON(data []byte) error {
	var raw pgConfigRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pg.Host = shared.ReplaceEnvVariables(raw.Host)
	pg.User = shared.ReplaceEnvVariables(raw.User)
	pg.Password = shared.ReplaceEnvVariables(raw.Password)
	pg.DbName = shared.ReplaceEnvVariables(raw.DbName)
	pg.Params = make(map[string]string)
	for key, value := range raw.Params {
		pg.Params[key] = shared.ReplaceEnvVariables(value)
	}
	// Handle Port which can be int or string
	switch v := raw.Port.(type) {
	case float64:
		pg.Port = int(v)
	case string:
		port, err := strconv.Atoi(shared.ReplaceEnvVariables(v))
		if err != nil {
			return fmt.Errorf("invalid port value: %v", err)
		}
		pg.Port = port
	default:
		return fmt.Errorf("invalid port type")
	}
	return nil
}
