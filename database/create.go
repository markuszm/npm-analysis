package database

import (
	"database/sql"
	"github.com/pkg/errors"
)

func CreateTables(db *sql.DB) error {
	createPackages := `
	CREATE TABLE IF NOT EXISTS packages(
		name VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL PRIMARY KEY,
		version VARCHAR(255),
		description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
		homepage TEXT,
		main TEXT,
		npmVersion VARCHAR(255),
		nodeVersion VARCHAR(255)		
	);
	`
	_, execErr := db.Exec(createPackages)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating package table")
	}

	createKeywords := `
	CREATE TABLE IF NOT EXISTS keywords(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createKeywords)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating keywords table")
	}

	createLicense := `
	CREATE TABLE IF NOT EXISTS license(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		type VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createLicense)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating license table")
	}

	createNpmUser := `
	CREATE TABLE IF NOT EXISTS npmUser(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		email VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createNpmUser)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating npmUser table")
	}

	createAuthors := `
	CREATE TABLE IF NOT EXISTS authors(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createAuthors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating authors table")
	}

	createContributors := `
	CREATE TABLE IF NOT EXISTS contributors(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createContributors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating contributors table")
	}

	createMaintainers := `
	CREATE TABLE IF NOT EXISTS maintainers(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createMaintainers)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating maintainers table")
	}

	createFiles := `
	CREATE TABLE IF NOT EXISTS files(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createFiles)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating files table")
	}

	// TODO: add directories table if necessary

	createRepository := `
	CREATE TABLE IF NOT EXISTS repository(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		type VARCHAR(255),
		url TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createRepository)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating repository table")
	}

	createScripts := `
	CREATE TABLE IF NOT EXISTS scripts(
		name VARCHAR(255),
		command TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createScripts)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating scripts table")
	}

	createBin := `
	CREATE TABLE IF NOT EXISTS bin(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		path TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createBin)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bin table")
	}

	createMan := `
	CREATE TABLE IF NOT EXISTS man(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createMan)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating man table")
	}

	createDependencies := `
	CREATE TABLE IF NOT EXISTS dependencies(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating dependencies table")
	}

	createDevDependencies := `
	CREATE TABLE IF NOT EXISTS devDependencies(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createDevDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating devDependencies table")
	}

	createPeerDependencies := `
	CREATE TABLE IF NOT EXISTS peerDependencies(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createPeerDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating peerDependencies table")
	}

	createBundledDependencies := `
	CREATE TABLE IF NOT EXISTS bundledDependencies(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createBundledDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bundledDependencies table")
	}

	createOptionalDependencies := `
	CREATE TABLE IF NOT EXISTS optionalDependencies(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createOptionalDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating optionalDependencies table")
	}

	createEngines := `
	CREATE TABLE IF NOT EXISTS engines(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createEngines)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating engines table")
	}

	createOs := `
	CREATE TABLE IF NOT EXISTS os(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createOs)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating os table")
	}

	createCpu := `
	CREATE TABLE IF NOT EXISTS cpu(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		name VARCHAR(255),
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createCpu)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating cpu table")
	}

	createDist := `
	CREATE TABLE IF NOT EXISTS dist(
		id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
		shasum VARCHAR(255),
		tarball TEXT,
		package VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(createDist)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating dist table")
	}

	// TODO: add publishConfig table if necessary

	return nil
}

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

func CreateMaintainersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS maintainersVersion(
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
		return errors.Wrap(err, "Error creating maintainersVersion table")
	}

	return nil
}
