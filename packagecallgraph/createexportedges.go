package packagecallgraph

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"sync"
)

type ExportEdgeCreator struct {
	neo4jUrl  string
	inputFile string
	logger    *zap.SugaredLogger
	workers   int
}

func NewExportEdgeCreator(neo4jUrl, inputFile string, workerNumber int, logger *zap.SugaredLogger) *ExportEdgeCreator {
	return &ExportEdgeCreator{neo4jUrl: neo4jUrl, inputFile: inputFile, workers: workerNumber, logger: logger}
}

func (e *ExportEdgeCreator) Exec() error {
	file, err := os.Open(e.inputFile)
	if err != nil {
		return errors.Wrap(err, "error opening export result file - does it exist?")
	}

	decoder := json.NewDecoder(file)

	workerWait := sync.WaitGroup{}

	jobs := make(chan model.PackageResult, 100)

	for w := 1; w <= e.workers; w++ {
		workerWait.Add(1)
		go e.worker(w, jobs, &workerWait)
	}

	for {
		result := model.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				e.logger.Debug("finished decoding result json")
				break
			} else {
				return errors.Wrap(err, "error processing package results")
			}
		}

		jobs <- result
	}

	close(jobs)
	workerWait.Wait()

	return err
}

func (e *ExportEdgeCreator) worker(workerId int, jobs chan model.PackageResult, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	err := neo4JDatabase.InitDB(e.neo4jUrl)
	if err != nil {
		e.logger.Fatal(err)
	}
	defer neo4JDatabase.Close()

	for j := range jobs {
		pkg := j.Name

		exports, err := resultprocessing.TransformToExports(j.Result)
		if err != nil {
			e.logger.Fatal(err)
		}

		for _, export := range exports {
			err := e.addExportEdges(pkg, export, neo4JDatabase)
			if err != nil {
				e.logger.Fatalw("error inserting export", "export", export, "error", err)
			}
		}

		e.logger.Debugf("Worker: %v, Package: %s, Exports %v", workerId, j.Name, len(exports))
	}
	workerWait.Done()
}

func (e *ExportEdgeCreator) addExportEdges(pkgName string, export resultprocessing.Export, database graph.Database) error {
	if export.ExportType != "function" {
		return nil
	}

	_, err := database.Exec(`
		MERGE (l:LocalFunction {name: {fullLocalFunctionName}, functionName: {fromFunction}})
		MERGE (e:ExportedFunction {name: {fullExportedFunctionName}, functionName: {exportedFunction}})
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (l)-[:EXPORT_AS]->(e)
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(e)
		MERGE (m)-[:CONTAINS_FUNCTION]->(l)
		`,
		map[string]interface{}{
			"packageName":              pkgName,
			"fullModuleName":           fmt.Sprintf("%s|%s", pkgName, export.File),
			"moduleName":               export.File,
			"fullLocalFunctionName":    fmt.Sprintf("%s|%s|%s", pkgName, export.File, e.getLocalName(export)),
			"fromFunction":             e.getLocalName(export),
			"fullExportedFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, export.File, export.Identifier),
			"exportedFunction":         export.Identifier,
		})
	if err != nil {
		return errors.Wrap(err, "error inserting module node")
	}

	return nil
}

func (e *ExportEdgeCreator) getLocalName(export resultprocessing.Export) string {
	if export.Local != "" {
		return export.Local
	}
	return export.Identifier
}
