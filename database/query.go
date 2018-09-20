package database

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"log"
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

func MainFileForPackage(db *sql.DB, packageName string) (string, error) {
	var mainFile sql.NullString
	rows, err := db.Query("SELECT main FROM packages WHERE name = ?", packageName)
	if err != nil {
		return mainFile.String, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&mainFile)
		if err != nil {
			return mainFile.String, err
		}
	}
	err = rows.Err()
	if err != nil {
		return mainFile.String, err
	}
	return mainFile.String, nil
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

func GetDependents(db *sql.DB, packageName string) ([]string, error) {
	var dependents []string

	rows, err := db.Query("SELECT package FROM dependencies WHERE name = ?", packageName)
	if err != nil {
		return dependents, errors.Wrap(err, "Failed to query dependents")
	}

	defer rows.Close()
	for rows.Next() {
		var dependent string

		err = rows.Scan(&dependent)
		if err != nil {
			return dependents, errors.Wrap(err, "Could not get info from row")
		}

		dependents = append(dependents, dependent)
	}

	err = rows.Err()
	if err != nil {
		return dependents, err
	}
	return dependents, nil
}

func GetPackages(db *sql.DB) ([]string, error) {
	var packages []string

	rows, err := db.Query("select name from packages where name <> \"\"")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query packages")
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
	return packages, err
}

func GetPackagesWithVersion(db *sql.DB) ([]model.PackageVersionPair, error) {
	var packages []model.PackageVersionPair

	rows, err := db.Query("select name, version from packages where name <> \"\"")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query packages")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			pkgName string
			version string
		)
		err := rows.Scan(&pkgName, &version)
		if err != nil {
			return packages, errors.Wrap(err, "Could not get info from row")
		}

		packages = append(packages, model.PackageVersionPair{Name: pkgName, Version: version})
	}
	err = rows.Err()
	return packages, err
}

type Person struct {
	Name, Email, Url, PackageName string
}

func GetAuthors(db *sql.DB) ([]Person, error) {
	var authors []Person

	rows, err := db.Query("select * from authors")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query dependencies")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id                        int
			name, email, url, pkgName string
		)
		err := rows.Scan(&id, &name, &email, &url, &pkgName)
		if err != nil {
			return authors, errors.Wrap(err, "Could not get info from row")
		}

		author := Person{
			Name:        name,
			Email:       email,
			Url:         url,
			PackageName: pkgName,
		}

		authors = append(authors, author)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return authors, nil
}

func GetMaintainers(db *sql.DB) ([]Person, error) {
	var maintainers []Person

	rows, err := db.Query("select * from maintainers")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query dependencies")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id                        int
			name, email, url, pkgName string
		)
		err := rows.Scan(&id, &name, &email, &url, &pkgName)
		if err != nil {
			return maintainers, errors.Wrap(err, "Could not get info from row")
		}

		author := Person{
			Name:        name,
			Email:       email,
			Url:         url,
			PackageName: pkgName,
		}

		maintainers = append(maintainers, author)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return maintainers, nil
}
