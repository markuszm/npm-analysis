package insert

import (
	"database/sql"
	"npm-analysis/database/model"
)

func StoreAuthor(db *sql.DB, pkg model.Package) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}

	execErr := insertAuthorField(pkg, tx)
	if execErr != nil {
		return execErr
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func insertAuthorField(pkg model.Package, tx *sql.Tx) error {
	queryInsertAuthor := `
		INSERT INTO authors(name, email, url, package) values(?,?,?,?)
	`

	pkgName := pkg.Name
	author := pkg.Author

	var err error

	if author == nil {
		return nil
	}
	switch author.(type) {
	case string:
		author = author.(string)
		_, err = tx.Exec(queryInsertAuthor, author, "", "", pkgName)
	case []interface{}:
		str := ""
		for _, v := range author.([]interface{}) {
			str += v.(string) + ","
		}
		author = str
		_, err = tx.Exec(queryInsertAuthor, author, "", "", pkgName)
	case map[string]interface{}:
		authorMap := author.(map[string]interface{})
		// todo: very dirty hack to avoid nil in map
		name := authorMap["name"]
		if name == nil {
			name = ""
		}

		email := authorMap["email"]
		if email == nil {
			email = ""
		}
		url := authorMap["url"]
		if url == nil {
			url = ""
		}
		person := model.Person{
			Name:  name.(string),
			Email: email.(string),
			Url:   url.(string),
		}
		_, err = tx.Exec(queryInsertAuthor, person.Name, person.Email, person.Url, pkgName)
	}
	return err
}
