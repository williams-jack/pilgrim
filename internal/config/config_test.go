package config

import (
	"strings"
	"testing"

	"github.com/williams-jack/pilgrim/internal/postgres"
	"github.com/williams-jack/pilgrim/internal/shared"
)

func TestReadFromReader_PostgresConfig(t *testing.T) {
	jsonConfig := `
	{
		"dbType": "postgres",
		"downDir": "migrations/down",
		"upDir": "migrations/up",
		"connectionString": "custom-connection-string",
		"dbConfig": {
			"host": "localhost",
			"port": 5432,
			"user": "testuser",
			"password": "testpass",
			"dbName": "testdb",
			"params": {
				"sslmode": "disable"
			}
		}
	}`

	reader := strings.NewReader(jsonConfig)
	cfg, err := ReadFromReader(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DbType != shared.DbType("postgres") {
		t.Errorf("expected DbType 'postgres', got '%v'", cfg.DbType)
	}
	if cfg.DownDir != "migrations/down" {
		t.Errorf("expected DownDir 'migrations/down', got '%v'", cfg.DownDir)
	}
	if cfg.UpDir != "migrations/up" {
		t.Errorf("expected UpDir 'migrations/up', got '%v'", cfg.UpDir)
	}
	if cfg.ConnectionString() != "custom-connection-string" {
		t.Errorf("expected ConnectionString 'custom-connection-string', got '%v'", cfg.ConnectionString())
	}

	pgCfg, ok := cfg.dbConfig.(*postgres.PgConfig)
	if !ok {
		t.Fatalf("dbConfig is not of type PgConfig")
	}
	if pgCfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got '%v'", pgCfg.Host)
	}
	if pgCfg.Port != 5432 {
		t.Errorf("expected Port 5432, got %v", pgCfg.Port)
	}
	if pgCfg.User != "testuser" {
		t.Errorf("expected User 'testuser', got '%v'", pgCfg.User)
	}
	if pgCfg.Password != "testpass" {
		t.Errorf("expected Password 'testpass', got '%v'", pgCfg.Password)
	}
	if pgCfg.DbName != "testdb" {
		t.Errorf("expected DbName 'testdb', got '%v'", pgCfg.DbName)
	}
	if pgCfg.Params["sslmode"] != "disable" {
		t.Errorf("expected Params['sslmode'] 'disable', got '%v'", pgCfg.Params["sslmode"])
	}
}

func TestReadFromReader_UnsupportedDbType(t *testing.T) {
	jsonConfig := `
	{
		"dbType": "mysql",
		"downDir": "migrations/down",
		"upDir": "migrations/up",
		"connectionString": "irrelevant",
		"dbConfig": {}
	}`

	reader := strings.NewReader(jsonConfig)
	_, err := ReadFromReader(reader)
	if err == nil {
		t.Fatal("expected error for unsupported dbType, got nil")
	}
}

func TestReadFromReader_EnvVariableReplacement(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("PG_HOST", "envhost")
	t.Setenv("PG_USER", "envuser")
	t.Setenv("PG_PASS", "envpass")
	t.Setenv("PG_DB", "envdb")
	t.Setenv("PG_PORT", "1234")
	t.Setenv("PG_SSLMODE", "require")
	t.Setenv("CUSTOM_CONN", "env-conn-string")

	jsonConfig := `
	{
		"dbType": "postgres",
		"downDir": "${PG_DB}/down",
		"upDir": "${PG_DB}/up",
		"connectionString": "${CUSTOM_CONN}",
		"dbConfig": {
			"host": "${PG_HOST}",
			"port": "${PG_PORT}",
			"user": "${PG_USER}",
			"password": "${PG_PASS}",
			"dbName": "${PG_DB}",
			"params": {
				"sslmode": "${PG_SSLMODE}"
			}
		}
	}`

	reader := strings.NewReader(jsonConfig)
	cfg, err := ReadFromReader(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DownDir != "envdb/down" {
		t.Errorf("expected DownDir 'envdb/down', got '%v'", cfg.DownDir)
	}
	if cfg.UpDir != "envdb/up" {
		t.Errorf("expected UpDir 'envdb/up', got '%v'", cfg.UpDir)
	}
	if cfg.ConnectionString() != "env-conn-string" {
		t.Errorf("expected ConnectionString 'env-conn-string', got '%v'", cfg.ConnectionString())
	}

	pgCfg, ok := cfg.dbConfig.(*postgres.PgConfig)
	if !ok {
		t.Fatalf("dbConfig is not of type PgConfig")
	}
	if pgCfg.Host != "envhost" {
		t.Errorf("expected Host 'envhost', got '%v'", pgCfg.Host)
	}
	if pgCfg.Port != 1234 {
		t.Errorf("expected Port 1234, got %v", pgCfg.Port)
	}
	if pgCfg.User != "envuser" {
		t.Errorf("expected User 'envuser', got '%v'", pgCfg.User)
	}
	if pgCfg.Password != "envpass" {
		t.Errorf("expected Password 'envpass', got '%v'", pgCfg.Password)
	}
	if pgCfg.DbName != "envdb" {
		t.Errorf("expected DbName 'envdb', got '%v'", pgCfg.DbName)
	}
	if pgCfg.Params["sslmode"] != "require" {
		t.Errorf("expected Params['sslmode'] 'require', got '%v'", pgCfg.Params["sslmode"])
	}
}
