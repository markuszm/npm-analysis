package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
)

type Mysql struct{}

func (d *Mysql) InitDB(dataSource string) (*sql.DB, error) {
	db, openError := sql.Open("mysql", dataSource)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxIdleConns(100)
	if openError != nil {
		return nil, errors.Wrap(openError, "Error opening database")
	}
	return db, nil
}
