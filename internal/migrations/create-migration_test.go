package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/williams-jack/pilgrim/internal/shared"
)

func TestCreateMigration_CreatesFilesAndDirs(t *testing.T) {
	tmpDir := t.TempDir()
	upDir := filepath.Join(tmpDir, "up")
	downDir := filepath.Join(tmpDir, "down")
	name := "test_migration"
	dbType := shared.DbTypePostgres

	err := CreateMigration(name, upDir, downDir, dbType, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if directories were created
	if _, err := os.Stat(upDir); os.IsNotExist(err) {
		t.Errorf("expected up directory to exist, but it does not")
	}
	if _, err := os.Stat(downDir); os.IsNotExist(err) {
		t.Errorf("expected down directory to exist, but it does not")
	}

	// Check if files were created
	files, err := os.ReadDir(upDir)
	if err != nil || len(files) == 0 {
		t.Errorf("expected migration file in up directory, but none found")
	}
	files, err = os.ReadDir(downDir)
	if err != nil || len(files) == 0 {
		t.Errorf("expected migration file in down directory, but none found")
	}
}

// Check migration files exist
func TestCreateMigration_DirsExist(t *testing.T) {
	tmpDir := t.TempDir()
	upDir := filepath.Join(tmpDir, "up")
	downDir := filepath.Join(tmpDir, "down")
	os.Mkdir(upDir, 0755)
	os.Mkdir(downDir, 0755)
	name := "already_exists"
	dbType := shared.DbTypePostgres

	err := CreateMigration(name, upDir, downDir, dbType, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if files were created
	files, err := os.ReadDir(upDir)
	if err != nil || len(files) == 0 {
		t.Errorf("expected migration file in up directory, but none found")
	}
	files, err = os.ReadDir(downDir)
	if err != nil || len(files) == 0 {
		t.Errorf("expected migration file in down directory, but none found")
	}
}

func TestCreateMigration_DirDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	upDir := filepath.Join(tmpDir, "up")
	downDir := filepath.Join(tmpDir, "down")
	name := "missing_dirs"
	dbType := shared.DbTypePostgres

	// Do not create dirs, expect error
	err := CreateMigration(name, upDir, downDir, dbType, false)
	if err == nil {
		t.Errorf("expected error when dirs do not exist, got nil")
	}
}

func TestCreateMigration_EmptyName(t *testing.T) {
	err := CreateMigration("   ", "up", "down", shared.DbTypePostgres, true)
	if err == nil {
		t.Errorf("expected error for empty migration name, got nil")
	}
}
