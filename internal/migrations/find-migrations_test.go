package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/williams-jack/pilgrim/internal/shared"
)

func createTestFile(t *testing.T, dir, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("failed to create file %s: %v", name, err)
	}
	f.Close()
}

func TestFindMigrations(t *testing.T) {
	tmpDir := t.TempDir()

	// Assume shared.DbTypePostgres has extension "sql"
	dbType := shared.DbTypePostgres

	// Valid migration files
	validFiles := []string{
		"20060102150405_init.sql",
		"20060102150405_add_table.sql",
	}
	// Invalid files
	invalidFiles := []string{
		"not_a_migration.txt",
		"20060102150405_wrong.ext",
		"20060102150405_missing.sqlx",
		"README.md",
	}

	for _, name := range validFiles {
		createTestFile(t, tmpDir, name)
	}
	for _, name := range invalidFiles {
		createTestFile(t, tmpDir, name)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	createTestFile(t, subDir, "20060102150405_sub.sql")

	files, err := FindMigrations(tmpDir, dbType)
	if err != nil {
		t.Fatalf("FindMigrations returned error: %v", err)
	}

	expected := map[string]bool{}
	for _, name := range validFiles {
		expected[filepath.Join(tmpDir, name)] = true
	}

	if len(files) != len(expected) {
		t.Errorf("expected %d migration files, got %d", len(expected), len(files))
	}
	for _, f := range files {
		if !expected[f] {
			t.Errorf("unexpected migration file found: %s", f)
		}
	}
}

func TestFindMigrations_NotADir(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "notadir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, err = FindMigrations(tmpFile.Name(), shared.DbTypePostgres)
	if err == nil {
		t.Error("expected error for non-directory path, got nil")
	}
}
