package packagecallgraph

import (
	"github.com/markuszm/npm-analysis/database/graph"
)

func Init(database graph.Database) error {
	_, err := database.Exec("CREATE CONSTRAINT ON (p:Package) ASSERT p.name IS UNIQUE", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE CONSTRAINT ON (m:Module) ASSERT m.name IS UNIQUE", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE CONSTRAINT ON (l:LocalFunction) ASSERT l.name IS UNIQUE", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE CONSTRAINT ON (e:ExportedFunction) ASSERT e.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE INDEX ON :Package(name)", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE INDEX ON :Module(name)", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE INDEX ON :LocalFunction(name)", nil)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE INDEX ON :ExportedFunction(name)", nil)
	if err != nil {
		return err
	}

	return nil
}
