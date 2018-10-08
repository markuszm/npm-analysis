package packagecallgraph

import (
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/markuszm/npm-analysis/database/graph"
)

type GraphQueries struct {
	neo4jUrl string
	db       graph.Database
}

func NewGraphQueries(neo4jUrl string) (*GraphQueries, error) {
	database := graph.NewNeo4JDatabase()
	err := database.InitDB(neo4jUrl)
	if err != nil {
		return nil, err
	}
	return &GraphQueries{neo4jUrl: neo4jUrl, db: database}, nil
}

func (q *GraphQueries) Close() {
	q.db.Close()
}

func (q *GraphQueries) StreamExportedFunctions(functionChan chan string) {
	database := graph.NewNeo4JDatabase()
	defer database.Close()
	err := database.InitDB(q.neo4jUrl)
	if err != nil {
		log.Fatal(err)
	}

	resultChan := make(chan []interface{}, 0)

	go database.QueryStream("MATCH (e:Function {functionType: \"export\"}) RETURN e.name", map[string]interface{}{}, resultChan)

	for r := range resultChan {
		// TODO: very unsafe - should check if result is valid
		functionChan <- r[0].(string)
	}

	close(functionChan)
}

func (q *GraphQueries) StreamPackages(packageChan chan string) {
	database := graph.NewNeo4JDatabase()
	defer database.Close()
	err := database.InitDB(q.neo4jUrl)
	if err != nil {
		log.Fatal(err)
	}

	resultChan := make(chan []interface{}, 0)

	go database.QueryStream("MATCH (p:Package) RETURN p.name", map[string]interface{}{}, resultChan)

	for r := range resultChan {
		// TODO: very unsafe - should check if result is valid
		packageChan <- r[0].(string)
	}

	close(packageChan)
}

func (q *GraphQueries) GetCallCountForExportedFunction(exportedFunction string) (int64, error) {
	result, err := q.db.Query("MATCH (e:Function)<-[:CALL]-(l:Function) WHERE e.name = {name} RETURN count(l)",
		map[string]interface{}{"name": exportedFunction})
	if err != nil {
		return 0, err
	}
	// TODO: very unsafe - should check if result is valid
	count := result[0][0].(int64)
	return count, nil
}

func (q *GraphQueries) GetPackagesThatCallExportedFunction(exportedFunction string) ([]string, error) {
	result, err := q.db.Query("MATCH (e:Function)<-[:CALL]-(:Function)<-[:CONTAINS_FUNCTION]-(:Module)<-[:CONTAINS_MODULE]-(p:Package) WHERE e.name = {name} RETURN DISTINCT p.name",
		map[string]interface{}{"name": exportedFunction})
	if err != nil {
		return nil, err
	}

	packages := make([]string, 0)
	for _, row := range result {
		// TODO: very unsafe - should check if result is valid
		pkg := row[0].(string)
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (q *GraphQueries) GetRequiredPackagesForPackage(packageName string) ([]string, error) {
	result, err := q.db.Query("MATCH (n:Package)-[:REQUIRES_PACKAGE]->(p:Package) WHERE n.name = {name} RETURN DISTINCT p.name",
		map[string]interface{}{"name": packageName})

	if err != nil {
		return nil, err
	}

	packages := make([]string, 0)
	for _, row := range result {
		// TODO: very unsafe - should check if result is valid
		pkg := row[0].(string)
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (q *GraphQueries) GetExportedFunctionsForPackage(packageName string, mainModuleName string) ([]string, error) {
	result, err := q.db.Query("MATCH (n:Package)-[:CONTAINS_MODULE]->(m:Module)-[:CONTAINS_FUNCTION]->(e:Function) "+
		"WHERE n.name = {name} AND m.name = {main} "+
		"WITH e MATCH (e)<-[c:CALL]-(:Function) WITH e, count(c) as calls "+
		"WHERE calls > 1 RETURN e.name", map[string]interface{}{"name": packageName, "main": mainModuleName})

	if err != nil {
		return nil, err
	}

	functions := make([]string, 0)
	for _, row := range result {
		// TODO: very unsafe - should check if result is valid
		pkg := row[0].(string)
		functions = append(functions, pkg)
	}

	return functions, nil
}

func (q *GraphQueries) GetFunctionsFromPackageThatCallAnotherFunctionDirectly(packageName, otherFunction string) ([]string, error) {
	result, err := q.db.Query("MATCH (f1:Function)<-[:CALL]-(pFunc:Function)<-[:CONTAINS_FUNCTION]-(:Module)<-[:CONTAINS_MODULE]-(p:Package) "+
		"WHERE f1.name = {functionName} AND p.name = {packageName} "+
		"RETURN pFunc.name", map[string]interface{}{"packageName": packageName, "functionName": otherFunction})

	if err != nil {
		return nil, err
	}

	functions := make([]string, 0)
	for _, row := range result {
		// TODO: very unsafe - should check if result is valid
		pkg := row[0].(string)
		functions = append(functions, pkg)
	}

	return functions, nil
}
