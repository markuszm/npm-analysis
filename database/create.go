package database

import (
	"database/sql"
	"github.com/pkg/errors"
)

func CreateTables(db *sql.DB) error {
	create_packages := `
	CREATE TABLE IF NOT EXISTS packages(
		name VARCHAR(255) NOT NULL PRIMARY KEY,
		version VARCHAR(255),
		description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
		homepage VARCHAR(255),
		main VARCHAR(255),
		npmVersion VARCHAR(255),
		nodeVersion VARCHAR(255)		
	);
	`
	_, execErr := db.Exec(create_packages)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating package table")
	}

	create_keywords := `
	CREATE TABLE IF NOT EXISTS keywords(
		name VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_keywords)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating keywords table")
	}

	create_license := `
	CREATE TABLE IF NOT EXISTS license(
		type VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_license)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating license table")
	}

	create_npmUser := `
	CREATE TABLE IF NOT EXISTS npmUser(
		name VARCHAR(255),
		email VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_npmUser)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating npmUser table")
	}

	create_authors := `
	CREATE TABLE IF NOT EXISTS authors(
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_authors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating authors table")
	}

	create_contributors := `
	CREATE TABLE IF NOT EXISTS contributors(
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_contributors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating contributors table")
	}

	create_maintainers := `
	CREATE TABLE IF NOT EXISTS maintainers(
		name VARCHAR(255),
		email VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_maintainers)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating maintainers table")
	}

	create_files := `
	CREATE TABLE IF NOT EXISTS files(
		name VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_files)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating files table")
	}

	// TODO: add directories table if necessary

	create_repository := `
	CREATE TABLE IF NOT EXISTS repository(
		type VARCHAR(255),
		url VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_repository)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating repository table")
	}

	create_scripts := `
	CREATE TABLE IF NOT EXISTS scripts(
		name VARCHAR(255),
		command VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_scripts)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating scripts table")
	}

	create_bin := `
	CREATE TABLE IF NOT EXISTS bin(
		name VARCHAR(255),
		path VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_bin)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bin table")
	}

	create_man := `
	CREATE TABLE IF NOT EXISTS man(
		name VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_man)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating man table")
	}

	create_dependencies := `
	CREATE TABLE IF NOT EXISTS dependencies(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_dependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating dependencies table")
	}

	create_devDependencies := `
	CREATE TABLE IF NOT EXISTS devDependencies(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_devDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating devDependencies table")
	}

	create_peerDependencies := `
	CREATE TABLE IF NOT EXISTS peerDependencies(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_peerDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating peerDependencies table")
	}

	create_bundledDependencies := `
	CREATE TABLE IF NOT EXISTS bundledDependencies(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_bundledDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bundledDependencies table")
	}

	create_optionalDependencies := `
	CREATE TABLE IF NOT EXISTS optionalDependencies(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_optionalDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating optionalDependencies table")
	}

	create_engines := `
	CREATE TABLE IF NOT EXISTS engines(
		name VARCHAR(255),
		version VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_engines)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating engines table")
	}

	create_os := `
	CREATE TABLE IF NOT EXISTS os(
		name VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_os)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating os table")
	}

	create_cpu := `
	CREATE TABLE IF NOT EXISTS cpu(
		name VARCHAR(255),
		package VARCHAR(255),
		FOREIGN KEY(package) REFERENCES packages(name)
	);
	`
	_, execErr = db.Exec(create_cpu)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating cpu table")
	}

	// TODO: add publishConfig table if necessary

	return nil
}
