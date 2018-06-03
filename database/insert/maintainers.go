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
	query := `INSERT INTO maintainerCount(name, count2010,count2011,count2012,count2013,count2014,count2015,count2016,count2017,count2018) values(?,?,?,?,?,?,?,?,?,?)`

	_, err := db.Exec(query, m.Name, m.Counts[2010], m.Counts[2011], m.Counts[2012], m.Counts[2013], m.Counts[2014], m.Counts[2015], m.Counts[2016], m.Counts[2017], m.Counts[2018])

	return err
}
