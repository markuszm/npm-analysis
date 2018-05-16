package graph

import "github.com/markuszm/npm-analysis/database/model"

func Init(database Database) error {
	_, err := database.Exec("CREATE CONSTRAINT ON (p:Package) ASSERT p.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE CONSTRAINT ON (a:Author) ASSERT a.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE CONSTRAINT ON (m:Maintainer) ASSERT m.name IS UNIQUE", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE INDEX ON :Package(name)", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE INDEX ON :Maintainer(name)", nil)
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE INDEX ON :Author(name)", nil)
	if err != nil {
		return err
	}
	return nil
}

func InsertPackage(database Database, name string) error {
	_, insertErr := database.Exec(`
		MERGE (p:Package {name: {p1}})`, map[string]interface{}{"p1": name})
	return insertErr
}

func InsertAuthorRelation(database Database, person model.Person, pkgName string) error {
	if person.Name == "" {
		return nil
	}
	_, insertErr := database.Exec(`
		MERGE (a:Author {name: {name}})
		MERGE (p:Package {name: {p}})
		MERGE (a)-[:CREATED]->(p)`, map[string]interface{}{"name": person.Name, "email": person.Email, "url": person.Url, "p": pkgName})
	return insertErr
}

func InsertMaintainerRelation(database Database, person model.Person, pkgName string) error {
	if person.Name == "" {
		return nil
	}
	_, insertErr := database.Exec(`
		MERGE (m:Maintainer {name: {name}})
		MERGE (p:Package {name: {p}})
		MERGE (m)-[:CREATED]->(p)`, map[string]interface{}{"name": person.Name, "email": person.Email, "url": person.Url, "p": pkgName})
	return insertErr
}

func InsertDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	return insertErr
}

func InsertDevDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_DEV {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	return insertErr

}

func InsertBundledDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_BUNDLED {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	return insertErr

}

func InsertOptionalDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_OPTIONAL {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	return insertErr

}

func InsertPeerDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_PEER {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	return insertErr
}
