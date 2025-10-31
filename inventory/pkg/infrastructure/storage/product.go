package storage

import (
	"database/sql"
	"fmt"
)

func NewStorage(databaseType string, user string, pwd string, dbname string) (*sql.DB, error) {
	source := fmt.Sprintf("%s:%s@/%s", user, pwd, dbname)
	db, err := sql.Open(databaseType, source)
	if err != nil {
		return nil, err
	}

	return db, nil
}
