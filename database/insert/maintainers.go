package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
)

func StoreMaintainers(db *sql.DB, pkg model.Package) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	queryInsertMaintainer := `
		INSERT INTO maintainers(name, email, url, package) values(?,?,?,?)
	`

	insertStmt, prepareErr := tx.Prepare(queryInsertMaintainer)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	maintainers := pkg.Maintainers
	if maintainers == nil {
		return nil
	}

	for _, maintainer := range maintainers {
		_, insertErr := insertStmt.Exec(maintainer.Name, maintainer.Email, maintainer.Url, pkg.Name)
		if insertErr != nil {
			return insertErr
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func StoreMaintainerChange(db *sql.DB, changes []evolution.MaintainerChange) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	queryInsertMaintainer := `
		INSERT INTO maintainerChanges(name, package, changeType, version, releaseTime) values(?,?,?,?,?)
	`

	insertStmt, prepareErr := tx.Prepare(queryInsertMaintainer)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	for _, c := range changes {
		_, insertErr := insertStmt.Exec(c.Name, c.PackageName, c.ChangeType, c.Version, c.ReleaseTime)
		if insertErr != nil {
			return insertErr
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func StoreMaintainerCount(db *sql.DB, m evolution.MaintainerCount) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	query := `INSERT INTO maintainerCount(name, year, month, count) values(?,?,?,?)`

	insertStmt, prepareErr := tx.Prepare(query)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	for year, monthMap := range m.Counts {
		for month, count := range monthMap {
			_, insertErr := insertStmt.Exec(m.Name, year, month, count)
			if insertErr != nil {
				return insertErr
			}
		}

	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}
