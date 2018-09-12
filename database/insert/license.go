package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/evolution"
	"time"
)

func StoreLicense(db *sql.DB, license License) error {
	queryInsertLicense := `
		INSERT INTO license(type, url, package) values(?,?,?)
	`

	_, err := db.Exec(queryInsertLicense, license.License, "", license.PkgName)

	if err != nil {
		return err
	}

	return nil
}

func StoreLicenseWithVersion(db *sql.DB, licenses []LicenseVersion) error {
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

func StoreLicenceChanges(db *sql.DB, changes []evolution.LicenseChange) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	query := `
		INSERT INTO licenseChange(package, version, licenseFROM, licenseTO, changeString, releaseTime) values(?,?,?,?,?,?)
	`

	insertStmt, prepareErr := tx.Prepare(query)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	for _, c := range changes {
		_, insertErr := insertStmt.Exec(c.PackageName, c.Version, c.LicenseFrom, c.LicenseTo, c.ChangeString, c.ReleaseTime)
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

type LicenseVersion struct {
	PkgName, License, Version string
	Time                      time.Time
}

type License struct {
	PkgName, License string
}
