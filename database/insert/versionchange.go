package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/evolution"
)

func StoreVersionChanges(db *sql.DB, changes []evolution.VersionChange) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	queryInsert := `
		INSERT INTO versionChanges(version, versionPrev, versionDiff, package, releaseTime) values(?,?,?,?,?)
	`

	insertStmt, prepareErr := tx.Prepare(queryInsert)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	for _, c := range changes {
		_, insertErr := insertStmt.Exec(c.Version, c.VersionPrev, c.VersionDiff, c.PackageName, c.ReleaseTime)
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
