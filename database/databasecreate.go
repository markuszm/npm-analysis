package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, openError := sql.Open("sqlite3", filepath)
	if openError != nil {
		return nil, errors.Wrap(openError, "Error opening database")
	}
	return db, nil
}

func CreateTables(db *sql.DB) error {
	create_packages := `
	CREATE TABLE IF NOT EXISTS packages(
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT,
		version TEXT,
		description TEXT,
		homepage TEXT,
		license TEXT, 
		main TEXT,
		_npmVersion TEXT,
		_nodeVersion TEXT		
	);
	`
	_, execErr := db.Exec(create_packages)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating package table")
	}

	create_keywords := `
	CREATE TABLE IF NOT EXISTS keywords(
		name TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_keywords)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating keywords table")
	}

	create_npmUser := `
	CREATE TABLE IF NOT EXISTS npmUser(
		name TEXT,
		email TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_npmUser)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating npmUser table")
	}

	create_authors := `
	CREATE TABLE IF NOT EXISTS authors(
		name TEXT,
		email TEXT,
		url TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_authors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating authors table")
	}

	create_contributors := `
	CREATE TABLE IF NOT EXISTS contributors(
		name TEXT,
		email TEXT,
		url TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_contributors)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating contributors table")
	}

	create_maintainers := `
	CREATE TABLE IF NOT EXISTS maintainers(
		name TEXT,
		email TEXT,
		url TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_maintainers)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating maintainers table")
	}

	create_files := `
	CREATE TABLE IF NOT EXISTS files(
		name TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_files)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating files table")
	}

	// TODO: add directories table if necessary

	create_repository := `
	CREATE TABLE IF NOT EXISTS repository(
		type TEXT,
		url TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_repository)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating repository table")
	}

	create_scripts := `
	CREATE TABLE IF NOT EXISTS scripts(
		name TEXT,
		command TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_scripts)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating scripts table")
	}

	create_bin := `
	CREATE TABLE IF NOT EXISTS bin(
		name TEXT,
		path TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_bin)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bin table")
	}

	create_man := `
	CREATE TABLE IF NOT EXISTS man(
		name TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_man)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating man table")
	}

	create_dependencies := `
	CREATE TABLE IF NOT EXISTS dependencies(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_dependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating dependencies table")
	}

	create_devDependencies := `
	CREATE TABLE IF NOT EXISTS devDependencies(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_devDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating devDependencies table")
	}

	create_peerDependencies := `
	CREATE TABLE IF NOT EXISTS peerDependencies(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_peerDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating peerDependencies table")
	}

	create_bundledDependencies := `
	CREATE TABLE IF NOT EXISTS bundledDependencies(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_bundledDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating bundledDependencies table")
	}

	create_optionalDependencies := `
	CREATE TABLE IF NOT EXISTS optionalDependencies(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_optionalDependencies)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating optionalDependencies table")
	}

	create_engines := `
	CREATE TABLE IF NOT EXISTS engines(
		name TEXT,
		version TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_engines)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating engines table")
	}

	create_os := `
	CREATE TABLE IF NOT EXISTS os(
		name TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_os)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating os table")
	}

	create_cpu := `
	CREATE TABLE IF NOT EXISTS cpu(
		name TEXT,
		package TEXT,
		FOREIGN KEY(package) REFERENCES packages(id)
	);
	`
	_, execErr = db.Exec(create_cpu)
	if execErr != nil {
		return errors.Wrap(execErr, "Error creating cpu table")
	}

	// TODO: add publishConfig table if necessary

	return nil
}