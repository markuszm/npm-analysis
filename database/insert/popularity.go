package insert

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/model"
)

func StorePopularity(popularity model.Popularity, db *sql.DB) error {
	query := `INSERT INTO popularity(package, overall, year2015, year2016, year2017, year2018) values(?,?,?,?,?,?)`

	_, err := db.Exec(query, popularity.PackageName, popularity.Overall, popularity.Year2015, popularity.Year2016, popularity.Year2017, popularity.Year2018)

	return err
}
