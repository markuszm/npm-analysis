package database

import "database/sql"

type Database interface {
	InitDB(dataSource string) (*sql.DB, error)
}
