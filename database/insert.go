package database

import (
	"database/sql"
	"npm-analysis/database/model"
)

func StorePackage(db *sql.DB, pkg model.Package) error {
	queryInsertPackage := `
	INSERT INTO packages(
		name,
		version,
		description,
		homepage,
		main,
		npmVersion,
		nodeVersion
	) values(?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE name = ?
	`

	_, execErr := db.Exec(queryInsertPackage, pkg.Name, pkg.Version, pkg.Description, pkg.Homepage, pkg.Main, pkg.NpmVersion, pkg.NodeVersion, pkg.Name)
	if execErr != nil {
		return execErr
	}

	return nil

}
