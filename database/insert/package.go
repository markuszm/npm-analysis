package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/model"
)

func StorePackage(db *sql.DB, pkg model.Package) error {
	queryInsertPackage := `
	INSERT INTO packages(
		name,
		version,
		description,
		homepage,
		main,
		npmVersion,
		nodeVersion
	) values(?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE name = ?;
	`

	main := handleMainField(pkg)

	_, execErr := db.Exec(queryInsertPackage, pkg.Name, pkg.Version, pkg.Description, pkg.Homepage, main, pkg.NpmVersion, pkg.NodeVersion, pkg.Name)
	if execErr != nil {
		return execErr
	}

	return nil
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
