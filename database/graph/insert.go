package graph

import "npm-analysis/database/model"

func InsertDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func InsertDevDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_DEV {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func InsertBundledDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_BUNDLED {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func InsertOptionalDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_OPTIONAL {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func InsertPeerDependency(neo4JDatabase Database, dep model.Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND_PEER {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.PkgName, "p2": dep.Name, "version": dep.Version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}
