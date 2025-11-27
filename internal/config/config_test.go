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

	pgCfg, ok := cfg.dbConfig.(postgres.PgConfig)
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
