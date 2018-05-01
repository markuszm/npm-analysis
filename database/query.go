package database

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"npm-analysis/database/model"
)

func FindPackage(db *sql.DB, packageName string) (string, error) {
	pkg := ""
	rows, err := db.Query("select name from packages where name = ?", packageName)
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

// unrolls the rows from db to an array to avoid timeouts on large number of rows
func GetDependencies(db *sql.DB, depType string) ([]model.Dependency, error) {
	var dependencies []model.Dependency

	rows, err := db.Query(fmt.Sprintf("select * from %s", depType))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query dependencies")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id                     int
			name, version, pkgName string
		)
		err := rows.Scan(&id, &name, &version, &pkgName)
		if err != nil {
			return dependencies, errors.Wrap(err, "Could not get info from row")
		}

		dep := model.Dependency{ID: id, PkgName: pkgName, Name: name, Version: version}
		dependencies = append(dependencies, dep)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return dependencies, nil
}

func GetPackages(db *sql.DB) ([]string, error) {
	var packages []string

	rows, err := db.Query("select name from packages")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query dependencies")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			pkgName string
		)
		err := rows.Scan(&pkgName)
		if err != nil {
			return packages, errors.Wrap(err, "Could not get info from row")
		}

		packages = append(packages, pkgName)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return packages, nil
}
