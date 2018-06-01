package database

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/pkg/errors"
	"log"
	"time"
)

func GetVersionChangesForPackage(pkg string, db *sql.DB) ([]evolution.VersionChange, error) {
	var changes []evolution.VersionChange

	rows, err := db.Query("select * from versionChanges where package=?", pkg)
	if err != nil {
		return changes, errors.Wrap(err, "Failed to query version changes")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id          int
			version     string
			versionPrev string
			versionDiff string
			pkg         string
			releaseTime string
		)
		err := rows.Scan(&id, &version, &versionPrev, &versionDiff, &pkg, &releaseTime)
		if err != nil {
			return changes, errors.Wrap(err, "Could not get info from row")
		}

		parsedTime, err := time.Parse("2006-01-02 15:04:05", releaseTime)
		if err != nil {
			parsedTime = time.Unix(1, 0)
		}

		changes = append(changes, evolution.VersionChange{
			PackageName: pkg,
			Version:     version,
			VersionPrev: versionPrev,
			VersionDiff: versionDiff,
			ReleaseTime: parsedTime,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return changes, nil
}
