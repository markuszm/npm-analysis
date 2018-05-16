package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/database/model"
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
