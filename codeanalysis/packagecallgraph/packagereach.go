package packagecallgraph

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/database"
	"log"
)

func PackageReach(pkg string, packages map[string]bool, db *sql.DB) {
	dependents, err := database.GetDependents(db, pkg)
	if err != nil {
		log.Fatalf("could not calculate package reach for package %s", pkg)
	}
	for _, dependent := range dependents {
		if ok := packages[dependent]; !ok {
			packages[dependent] = true
			PackageReach(dependent, packages, db)
		}
	}
}

func PackageReachLayer(pkg string, packages map[string]ReachDetails, db *sql.DB, layer int) {
	dependents, err := database.GetDependents(db, pkg)
	if err != nil {
		log.Fatalf("could not calculate package reach for package %s", pkg)
	}
	for _, dependent := range dependents {
		if _, ok := packages[dependent]; !ok {
			packages[dependent] = ReachDetails{layer, pkg}
			PackageReachLayer(dependent, packages, db, layer+1)
		}
	}
}

type ReachDetails struct {
	Layer      int
	Dependency string
}
