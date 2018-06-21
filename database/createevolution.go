package database

import (
	"database/sql"
	"github.com/pkg/errors"
)

func CreateLicenseTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS licenseVersion(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		version TEXT,
		releaseTime TIMESTAMP, 
		license TEXT,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating licenseVersion table")
	}

	return nil
}

func CreateLicenseChangeTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS licenseChange(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		version TEXT,
		licenseFROM TEXT,
		licenseTO TEXT,
		changeString TEXT,
		releaseTime TIMESTAMP, 
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating licenseVersion table")
	}

	return nil
}

func CreateMaintainerChangeTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS maintainerChanges(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		changeType VARCHAR(255),
		version TEXT,
		releaseTime TIMESTAMP, 
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating maintainersChange table")
	}

	return nil
}

func CreateDependencyChangeTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS dependencyChanges(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		dependency VARCHAR(255),
		depVersion VARCHAR(255),
		depVersionPrev VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		changeType VARCHAR(255),
		version VARCHAR(255),
		releaseTime TIMESTAMP, 
		FOREIGN KEY(package) REFERENCES packages(name)
	);	
	`

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating dependencyChange table")
	}

	return nil
}

func CreateVersionChangeTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS versionChanges(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		version VARCHAR(255),
		versionPrev VARCHAR(255),
		versionDiff VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		timeDiff DOUBLE,
		releaseTime TIMESTAMP,
		FOREIGN KEY(package) REFERENCES packages(name),
    	INDEX versionDiffIndex (versionDiff)
	);	
	`

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating versionChange table")
	}

	return nil
}

func CreatePopularity(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS popularity(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		overall DOUBLE,
		year2015 DOUBLE,
		year2016 DOUBLE,
		year2017 DOUBLE,
		year2018 DOUBLE,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating popularity table")
	}

	return nil
}

func CreateVersionCount(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS versionCount(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		major INT,
		minor INT,
		patch INT,
		avgMinorBetweenMajor DOUBLE,
		avgPatchBetweenMajor DOUBLE,
		avgPatchBetweenMinor DOUBLE,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating version count table")
	}

	return nil
}

func CreateMaintainerCount(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS maintainerCount(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255),
		year int,
		month int,
		count int,
		INDEX maintainerIndex (year,month,count)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "Error creating maintainer count table")
	}

	return nil
}
