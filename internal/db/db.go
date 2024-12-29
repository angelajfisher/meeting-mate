package db

import (
	"fmt"

	"zombiezen.com/go/sqlite/sqlitex"
)

type DatabasePool struct {
	location string // cleaned filepath to db
	pool     *sqlitex.Pool
}

func NewDatabasePool(cleanedDbPath string) (DatabasePool, error) {
	pool, err := sqlitex.NewPool(cleanedDbPath, sqlitex.PoolOptions{})
	if err != nil {
		return DatabasePool{}, fmt.Errorf("failed to create new pool: %w", err)
	}

	return DatabasePool{
			location: cleanedDbPath,
			pool:     pool,
		},
		nil
}
