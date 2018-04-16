package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Sqlite struct{}

func (d *Sqlite) InitDB(dataSource string) (*sql.DB, error) {
	db, openError := sql.Open("sqlite3", dataSource)
	if openError != nil {
		return nil, errors.Wrap(openError, "Error opening database")
	}
	return db, nil
}
