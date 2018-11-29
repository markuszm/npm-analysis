package packagecallgraph

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/database"
	"log"
)

func PackageCost(pkg string, packages map[string]bool, db *sql.DB) {
	dependencies, err := database.GetDependencies(db, pkg)
	if err != nil {
		log.Fatalf("could not calculate package cost for package %s with error: %s", pkg, err)
	}
	for _, dependency := range dependencies {
		if ok := packages[dependency]; !ok {
			packages[dependency] = true
			PackageCost(dependency, packages, db)
		}
	}
}
