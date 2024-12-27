package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
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

// Updates database schema as needed
func MakeMigrations(path string) error {
	schema := []string{`
		CREATE TABLE IF NOT EXISTS meetings (
			meeting_id TEXT PRIMARY KEY,
			name TEXT
		);
	`, `
		CREATE TABLE IF NOT EXISTS history_types (
			type TEXT PRIMARY KEY
		);
	`, `
		INSERT INTO history_types (type)
		VALUES
			('Full'),
			('Partial'),
			('Minimal');
	`, `
		CREATE TABLE IF NOT EXISTS watches (
			meeting_id TEXT NOT NULL,
			server_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			silent BOOL DEFAULT 1,
			summary BOOL DEFAULT 1,
			history_type TEXT NOT NULL DEFAULT 'Partial',
			command TEXT NOT NULL,
			link TEXT,
			PRIMARY KEY(meeting_id, server_id),
			FOREIGN KEY (meeting_id)
				REFERENCES meetings (meeting_id),
			FOREIGN KEY (history_type)
				REFERENCES history_types (type)
		);
	`}

	pool := sqlitemigration.NewPool(
		filepath.Clean(path),
		sqlitemigration.Schema{
			Migrations: schema,
			MigrationOptions: []*sqlitemigration.MigrationOptions{
				{
					DisableForeignKeys: false,
				},
			},
		},
		sqlitemigration.Options{
			Flags: sqlite.OpenReadWrite | sqlite.OpenCreate,
			PrepareConn: func(conn *sqlite.Conn) error {
				// Enable foreign keys
				return sqlitex.ExecuteTransient(conn, "PRAGMA foreign_keys = ON;", nil)
			},
			OnError: func(e error) {
				log.Println("could not make database migrations: ", e)
			},
		})
	defer pool.Close()

	// Migrations are blocking, so use a new connection as an indicator for their completion before closing the pool
	conn, err := pool.Get(context.TODO())
	if err != nil {
		return fmt.Errorf("could not open connection to database: %w", err)
	}
	pool.Put(conn)

	return nil
}
