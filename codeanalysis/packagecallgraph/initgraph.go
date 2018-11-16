package packagecallgraph

import (
	"github.com/markuszm/npm-analysis/database/graph"
)

func InitSchema(neo4jUrl string) error {
	database := graph.NewNeo4JDatabase()
	err := database.InitDB(neo4jUrl)
	if err != nil {
		return err
	}
	defer database.Close()

	_, err = database.Exec("CREATE CONSTRAINT ON (p:Package) ASSERT p.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE CONSTRAINT ON (m:Module) ASSERT m.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE CONSTRAINT ON (c:Class) ASSERT c.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE CONSTRAINT ON (f:Function) ASSERT f.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	return nil
}
