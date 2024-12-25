package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"zombiezen.com/go/sqlite"
)

// Checks for an existing SQLite database at the given path and creates one if it does not already exist
func InitializeDatabase(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not get info on file '%s': %w", path, err)
	}

	// create intermediate folders
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create intermediate folders: %w", err)
	}

	// create the new database file
	conn, err := sqlite.OpenConn(path, sqlite.OpenReadWrite, sqlite.OpenCreate)
	if err != nil {
		return fmt.Errorf("could not create new database file: %w", err)
	}
	conn.Close()

	return nil
}
