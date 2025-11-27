package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/williams-jack/pilgrim/internal/shared"
)

func migrationFileRegex(dbType shared.DbType) (*regexp.Regexp, error) {
	extension := dbType.FileExtension()
	if extension == "" {
		return nil, shared.UnsupportedDbTypeError(dbType)
	}
	pattern := `^\d{14}_.+\.` + regexp.QuoteMeta(extension) + `$`
	return regexp.MustCompile(pattern), nil
}

func migrationNameRegex (dbType shared.DbType) (*regexp.Regexp, error) {
	extension := dbType.FileExtension()
	if extension == "" {
		return nil, shared.UnsupportedDbTypeError(dbType)
	}
	pattern := `^\d{14}_(.+)\.` + regexp.QuoteMeta(extension) + `$`
	return regexp.MustCompile(pattern), nil
}

// MigrationNameFromPath extracts the migration name from the given file path
// based on the naming convention for the specified database type.
// It returns the migration name without the numeric prefix and file extension.
// If the file name does not match the expected pattern, an empty string is returned.
func MigrationNameFromPath(path string, dbType shared.DbType) string {
	re, err := migrationNameRegex(dbType)
	if err != nil {
		return ""
	}
	base := filepath.Base(path)
	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// FindMigrations scans the specified directory for migration files
// that match the naming convention for the given database type.
// It returns a slice of file paths to the found migration files.
// Subdirectories and symlinks are ignored.
func FindMigrations(dir string, dbType shared.DbType) ([]string, error) {
	re, err := migrationFileRegex(dbType)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", dir)
	}
	migrationFiles := make([]string, 0)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if re.MatchString(entry.Name()) {
			migrationFiles = append(migrationFiles, filepath.Join(dir, entry.Name()))
		}
	}
	return migrationFiles, nil
}
