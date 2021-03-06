package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
)

func StoreDependencies(db *sql.DB, pkg model.Package) error {
	pkgName := pkg.Name

	queryInsertDependencies := `
		INSERT INTO dependencies(name, version, package) values(?,?,?)
	`
	queryInsertDevDependencies := `
		INSERT INTO devDependencies(name, version, package) values(?,?,?)
	`
	queryInsertPeerDependencies := `
		INSERT INTO peerDependencies(name, version, package) values(?,?,?)
	`
	//queryInsertBundledDependencies := `
	//	INSERT INTO bundledDependencies(name, version, package) values(?,?,?)
	//`

	queryInsertOptionalDependencies := `
		INSERT INTO optionalDependencies(name, version, package) values(?,?,?)
	`

	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	insertDepStmt, insertErr := insertDependencies(tx, queryInsertDependencies, pkgName, pkg.Dependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertDepStmt.Close()

	insertDevDepStmt, insertErr := insertDependencies(tx, queryInsertDevDependencies, pkgName, pkg.DevDependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertDevDepStmt.Close()

	insertPeerDepStmt, insertErr := insertDependencies(tx, queryInsertPeerDependencies, pkgName, pkg.PeerDependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertPeerDepStmt.Close()

	//insertBundledDepStmt, insertErr := insertDependencies(tx, queryInsertBundledDependencies, pkgName, pkg.BundledDependencies)
	//if insertErr != nil {
	//	return insertErr
	//}
	//defer insertBundledDepStmt.Close()

	insertOptDepStmt, insertErr := insertDependencies(tx, queryInsertOptionalDependencies, pkgName, pkg.OptionalDependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertOptDepStmt.Close()

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func insertDependencies(tx *sql.Tx, query, pkgName string, deps map[string]string) (*sql.Stmt, error) {
	insertDepStmt, prepareErr := tx.Prepare(query)
	if prepareErr != nil {
		return nil, prepareErr
	}

	for n, v := range deps {
		_, execErr := insertDepStmt.Exec(n, v, pkgName)
		if execErr != nil {
			return nil, execErr
		}
	}
	return insertDepStmt, nil
}

func StoreDependencyChanges(db *sql.DB, changes []evolution.DependencyChange) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	queryInsert := `
		INSERT INTO dependencyChanges(dependency, depVersion, depVersionPrev, package, changeType, version,releaseTime) values(?,?,?,?,?,?,?)
	`

	insertStmt, prepareErr := tx.Prepare(queryInsert)
	if prepareErr != nil {
		return prepareErr
	}
	defer insertStmt.Close()

	for _, c := range changes {
		_, insertErr := insertStmt.Exec(c.DependencyName, c.DependencyVersion, c.DependencyVersionPrev, c.PackageName, c.ChangeType, c.PackageVersion, c.ReleaseTime)
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
