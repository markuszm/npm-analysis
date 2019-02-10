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

func PackageReachDev(pkg string, packages map[string]bool, db *sql.DB) {
	DirectDevReach(pkg, packages, db)
	dependents, err := database.GetDependents(db, pkg)
	if err != nil {
		log.Fatalf("could not calculate package reach for package %s", pkg)
	}
	for _, dependent := range dependents {
		if ok := packages[dependent]; !ok {
			packages[dependent] = true
			PackageReachDev(dependent, packages, db)
		}
	}
}

func DirectDevReach(pkg string, packages map[string]bool, db *sql.DB) {
	devDependents, err := database.GetDevDependents(db, pkg)
	if err != nil {
		log.Fatalf("could not get dev dependents for package %s", pkg)
	}
	for _, devDependent := range devDependents {
		packages[devDependent] = true
	}
}

type ReachDetails struct {
	Layer      int
	Dependency string
}
