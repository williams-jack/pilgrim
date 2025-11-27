package postgres

import (
	"strings"
	"testing"
)

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config PgConfig
		want   string
	}{
		{
			name: "No password",
			config: PgConfig{
				Host:   "db.example.com",
				Port:   5432,
				User:   "admin",
				DbName: "prod",
				Params: map[string]string{},
			},
			want: "postgres://admin@db.example.com:5432/prod",
		},
		{
			name: "No user and password",
			config: PgConfig{
				Host:   "127.0.0.1",
				Port:   5432,
				DbName: "testdb",
			},
			want: "postgres://127.0.0.1:5432/testdb",
		},
		{
			name: "No port",
			config: PgConfig{
				Host:   "localhost",
				User:   "user",
				Password: "pass",
				DbName: "db",
			},
			want: "postgres://user:pass@localhost/db",
		},
		{
			name: "No db name",
			config: PgConfig{
				Host: "localhost",
				Port: 5432,
			},
			want: "postgres://localhost:5432",
		},
		{
			name: "Empty config",
			config: PgConfig{},
			want: "postgres://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ConnectionString()
			if got != tt.want {
				t.Errorf("ConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}

	// Separated due to unordered map iteration
	t.Run("With params", func(t *testing.T) {
		config := PgConfig{
			Host:   "db.example.com",
			Port:   5432,
			User:   "admin",
			Password: "secret",
			DbName: "prod",
			Params: map[string]string{
				"sslmode": "disable",
				"pool":    "5",
			},
		}
		got := config.ConnectionString()
		expected_prefix := "postgres://admin:secret@db.example.com:5432/prod?"
		if len(got) < len(expected_prefix) || got[:len(expected_prefix)] != expected_prefix {
			t.Errorf("ConnectionString() = %q, want prefix %q", got, expected_prefix)
		}
		param_str := got[len(expected_prefix):]
		params := map[string]bool{
			"sslmode=disable": false,
			"pool=5":          false,
		}
		for p := range strings.SplitSeq(param_str, "&") {
			if _, exists := params[p]; exists {
				params[p] = true
			} else {
				t.Errorf("Unexpected param %q in connection string", p)
			}
		}
		for p, found := range params {
			if !found {
				t.Errorf("Expected param %q not found in connection string", p)
			}
		}
	})
}
