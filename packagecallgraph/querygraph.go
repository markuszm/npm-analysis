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

	go database.QueryStream("MATCH (e:ExportedFunction) RETURN e.name", map[string]interface{}{}, resultChan)

	for r := range resultChan {
		// TODO: very unsafe - should check if result is valid
		functionChan <- r[0].(string)
	}

	close(functionChan)
}

func (q *GraphQueries) GetCallCountForExportedFunction(exportedFunction string) (int64, error) {
	result, err := q.db.Query("MATCH (e:ExportedFunction)<-[:CALL]-(l:LocalFunction) WHERE e.name = {name} RETURN count(l)",
		map[string]interface{}{"name": exportedFunction})
	if err != nil {
		return 0, err
	}
	// TODO: very unsafe - should check if result is valid
	count := result[0][0].(int64)
	return count, nil
}

func (q *GraphQueries) GetPackagesThatCallExportedFunction(exportedFunction string) ([]string, error) {
	result, err := q.db.Query("MATCH (e:ExportedFunction)-[:CALL]-(:LocalFunction)-[:CONTAINS_FUNCTION]-(:Module)-[:CONTAINS_MODULE]-(p:Package) WHERE e.name = {name} RETURN DISTINCT p.name",
		map[string]interface{}{"name": exportedFunction})
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, row := range result {
		// TODO: very unsafe - should check if result is valid
		pkg := row[0].(string)
		packages = append(packages, pkg)
	}

	return packages, nil
}
