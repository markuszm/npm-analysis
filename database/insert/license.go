package insert

import (
	"database/sql"
	"time"
)

func StoreLicenseWithVersion(db *sql.DB, licenses []License) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	queryInsertLicense := `
		INSERT INTO licenseVersion(package, version, releaseTime, license) values(?,?,?,?)
	`

	for _, l := range licenses {
		t := l.Time.Add(time.Second * 1)
		_, execErr := tx.Exec(queryInsertLicense, l.PkgName, l.Version, t, l.License)
		if execErr != nil {
			return execErr
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}

type License struct {
	PkgName, License, Version string
	Time                      time.Time
}
