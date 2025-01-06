package db

import (
	"fmt"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const connTimeout = 5 * time.Second

type DatabasePool struct {
	Enabled  bool   // whether the program should use an external database
	location string // cleaned filepath to db
	pool     *sqlitex.Pool
}

func NewDatabasePool(cleanedDbPath string) (DatabasePool, error) {
	pool, err := sqlitex.NewPool(cleanedDbPath, sqlitex.PoolOptions{PrepareConn: func(conn *sqlite.Conn) error {
		// Enable foreign keys
		return sqlitex.ExecuteTransient(conn, "PRAGMA foreign_keys = ON;", nil)
	}})
	if err != nil {
		return DatabasePool{}, fmt.Errorf("failed to create new pool: %w", err)
	}

	return DatabasePool{
			Enabled:  true,
			location: cleanedDbPath,
			pool:     pool,
		},
		nil
}
