package database

import (
	"database/sql"
	"npm-analysis/database/model"
)

func StorePackage(db *sql.DB, pkg model.Package) error {
	pkgName := pkg.Name

	queryInsertPackage := `
	REPLACE INTO packages(
		name,
		version,
		description,
		homepage,
		main,
		npmVersion,
		nodeVersion
	) values(?, ?, ?, ?, ?, ?, ?)
	`

	queryInsertDependencies := `
		INSERT INTO dependencies(name, version, package) values(?,?,?)
	`
	queryInsertDevDependencies := `
		INSERT INTO devDependencies(name, version, package) values(?,?,?)
	`
	queryInsertPeerDependencies := `
		INSERT INTO peerDependencies(name, version, package) values(?,?,?)
	`
	queryInsertBundledDependencies := `
		INSERT INTO bundledDependencies(name, version, package) values(?,?,?)
	`
	queryInsertOptionalDependencies := `
		INSERT INTO optionalDependencies(name, version, package) values(?,?,?)
	`

	queryInsertDist := `
		INSERT INTO dist(shasum, tarball, package) VALUES (?,?,?)
	`

	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	main := handleMainField(pkg)

	_, execErr := tx.Exec(queryInsertPackage, pkg.Name, pkg.Version, pkg.Description, pkg.Homepage, main, pkg.NpmVersion, pkg.NodeVersion)
	if execErr != nil {
		return execErr
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

	insertBundledDepStmt, insertErr := insertDependencies(tx, queryInsertBundledDependencies, pkgName, pkg.BundledDependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertBundledDepStmt.Close()

	insertOptDepStmt, insertErr := insertDependencies(tx, queryInsertOptionalDependencies, pkgName, pkg.OptionalDependencies)
	if insertErr != nil {
		return insertErr
	}
	defer insertOptDepStmt.Close()

	_, execErr = tx.Exec(queryInsertDist, pkg.Distribution.Shasum, pkg.Distribution.Tarball, pkgName)
	if execErr != nil {
		return execErr
	}

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

func handleMainField(pkg model.Package) interface{} {
	main := pkg.Main
	switch main.(type) {
	case string:
		main = main.(string)
	case []interface{}:
		str := ""
		for _, v := range main.([]interface{}) {
			str += v.(string) + ","
		}
		main = str
	case map[string]interface{}:
		str := ""
		for k, v := range main.(map[string]interface{}) {
			str += "Key: " + k + " Value: " + v.(string) + " , "
		}
		main = str
	}
	return main
}
