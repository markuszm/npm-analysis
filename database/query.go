package database

import (
	"database/sql"
	"npm-analysis/database/model"
)

func FindPackage(db *sql.DB, packageName string) (model.Package, error) {
	var pkg model.Package
	rows, err := db.Query("select * from packages where name = ?", packageName)
	if err != nil {
		return pkg, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&pkg)
		if err != nil {
			return pkg, err
		}
	}
	err = rows.Err()
	if err != nil {
		return pkg, err
	}
	return pkg, nil
}
