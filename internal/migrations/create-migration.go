package migrations

import (
	"errors"
	"os"
	"path"
	"strings"
	"time"

	"github.com/williams-jack/pilgrim/internal/shared"
)

// CreateMigration generates "up" and "down" migration files with the specified
// name (trimmed of whitespace on left and right) in the form of YYYYMMDDHHMMSS_name.ext
// where "ext" is determined by the database type. The files are created in the specified
// directories.
func CreateMigration(name, upDir, downDir string, dbType shared.DbType, createDirs bool) error {
	migrationName := strings.TrimSpace(name)
	if migrationName == "" {
		return errors.New("migration name cannot be empty")
	}
	if createDirs {
		err := os.MkdirAll(upDir, os.ModePerm)
		if err != nil {
			return err
		}
		err = os.MkdirAll(downDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(upDir); os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(downDir); os.IsNotExist(err) {
		return err
	}

	fileName := time.Now().Format("20060102150405") + "_" + migrationName + "." + dbType.FileExtension()
	upFilePath := path.Join(upDir, fileName)
	downFilePath := path.Join(downDir, fileName)

	upFile, err := os.Create(upFilePath)
	if err != nil {
		return err
	}
	defer upFile.Close()

	downFile, err := os.Create(downFilePath)
	if err != nil {
		return err
	}
	defer downFile.Close()

	return nil
}
