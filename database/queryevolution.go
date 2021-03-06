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
			timeDiff    float64
			releaseTime string
		)
		err := rows.Scan(&id, &version, &versionPrev, &versionDiff, &pkg, &timeDiff, &releaseTime)
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
			TimeDiff:    timeDiff,
			ReleaseTime: parsedTime,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return changes, nil
}

func GetMaintainerChanges(db *sql.DB) ([]evolution.MaintainerChange, error) {
	var changes []evolution.MaintainerChange

	rows, err := db.Query("SELECT * FROM maintainerChanges ORDER BY package, releaseTime")
	if err != nil {
		return changes, errors.Wrap(err, "Failed to query version changes")
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id          int
			name        string
			pkgName     string
			changeType  string
			version     string
			releaseTime string
		)
		err := rows.Scan(&id, &name, &pkgName, &changeType, &version, &releaseTime)
		if err != nil {
			return changes, errors.Wrap(err, "Could not get info from row")
		}

		parsedTime, err := time.Parse("2006-01-02 15:04:05", releaseTime)
		if err != nil {
			parsedTime = time.Unix(1, 0)
		}

		changes = append(changes, evolution.MaintainerChange{
			PackageName: pkgName,
			Name:        name,
			ChangeType:  changeType,
			Version:     version,
			ReleaseTime: parsedTime,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return changes, nil
}

func GetActiveMaintainersForMonth(year, month int, db *sql.DB) ([]string, error) {
	var maintainers []string

	rows, err := db.Query("SELECT name FROM maintainerCount WHERE year = ? AND month = ? and count > 0", year, month)
	if err != nil {
		return maintainers, errors.Wrap(err, "Failed to query maintainers")
	}
	defer rows.Close()

	for rows.Next() {
		var name string

		err := rows.Scan(&name)
		if err != nil {
			return maintainers, errors.Wrap(err, "Could not get info from row")
		}

		maintainers = append(maintainers, name)
	}
	return maintainers, nil
}

func GetMaintainerCountsForMaintainer(maintainerName string, db *sql.DB) (evolution.MaintainerCount, error) {
	var maintainerCount evolution.MaintainerCount
	maintainerCount.Name = maintainerName

	tempMap := evolution.CreateCountMap()

	rows, err := db.Query("SELECT year, month, count FROM maintainerCount WHERE name = ?", maintainerName)
	if err != nil {
		return maintainerCount, errors.Wrap(err, "Failed to query maintainers")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			year, month, count int
		)
		err := rows.Scan(&year, &month, &count)
		if err != nil {
			return maintainerCount, errors.Wrap(err, "Could not get info from row")
		}
		tempMap[year][month] = count
	}

	maintainerCount.Counts = tempMap
	return maintainerCount, nil
}

func GetMaintainerNames(db *sql.DB) ([]string, error) {
	var maintainerNames []string

	rows, err := db.Query("SELECT distinct name FROM maintainerCount WHERE name <> \"\" order by name")
	if err != nil {
		return maintainerNames, errors.Wrap(err, "Failed to query maintainerCount")
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return maintainerNames, errors.Wrap(err, "Could not get info from row")
		}
		maintainerNames = append(maintainerNames, name)
	}
	return maintainerNames, nil
}
