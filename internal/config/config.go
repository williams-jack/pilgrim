package config

import (
	"encoding/json"
	"errors"
	"github.com/williams-jack/pilgrim/internal/postgres"
	"github.com/williams-jack/pilgrim/internal/shared"
	"io"
)

// PilgrimConfig holds the overall configuration for Pilgrim, including
// database type, migration directories, connection string, and database-specific config.
type PilgrimConfig struct {
	DbType           shared.DbType
	DownDir          string
	UpDir            string
	dbConfig         DbConfig
	connectionString string
}

// rawPilgrimConfig is an internal struct used for initial JSON unmarshalling
type rawPilgrimConfig struct {
	DbType           shared.DbType   `json:"dbType"`
	DownDir          string          `json:"downDir"`
	UpDir            string          `json:"upDir"`
	ConnectionString string          `json:"connectionString"`
	RawDbConfig      json.RawMessage `json:"dbConfig"`
}

// ReadFromReader reads the Pilgrim configuration from an io.Reader
func ReadFromReader(reader io.Reader) (*PilgrimConfig, error) {
	decoder := json.NewDecoder(reader)
	var rawConfig rawPilgrimConfig
	if err := decoder.Decode(&rawConfig); err != nil {
		return nil, err
	}
	var dbConfig DbConfig
	switch rawConfig.DbType {
	case "postgres":
		var pgConfig postgres.PgConfig
		pgConfig.UnmarshalJSON(rawConfig.RawDbConfig)
		dbConfig = &pgConfig
	default:
		return nil, errors.New("unsupported dbType: " + string(rawConfig.DbType))
	}
	return &PilgrimConfig{
		DbType:           rawConfig.DbType,
		DownDir:          shared.ReplaceEnvVariables(rawConfig.DownDir),
		UpDir:            shared.ReplaceEnvVariables(rawConfig.UpDir),
		dbConfig:         dbConfig,
		connectionString: shared.ReplaceEnvVariables(rawConfig.ConnectionString),
	}, nil
}

// ConnectionString returns the database connection string.
func (pc PilgrimConfig) ConnectionString() string {
	if pc.connectionString != "" {
		return pc.connectionString
	}
	if pc.dbConfig != nil {
		return pc.dbConfig.ConnectionString()
	}
	return ""
}
