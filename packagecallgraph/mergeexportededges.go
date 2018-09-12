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
			e.logger.With("package", j.Name).Error(err)
		}

		retry := 0

		for _, export := range exports {
			err := e.addExportEdges(pkg, export, neo4JDatabase)

			if err != nil {
				for retry < 3 && err != nil {
					err = e.addExportEdges(pkg, export, neo4JDatabase)
					retry++
				}
				if err != nil {
					e.logger.With("package", pkg, "error", err).Error("error merging exports")
				}
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

	exportName := export.Identifier
	moduleName := export.File
	localName := e.getLocalName(export)

	if localName == exportName {
		return nil
	}

	_, err := database.Exec(`
		MATCH (l:Function {name: {localFullName}})
		MERGE (e:Function {name: {exportFullName}}) ON CREATE SET e.functionName = {exportName}, e.functionType = "export"
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(e)
		WITH [e,l] as nodes CALL apoc.refactor.mergeNodes(nodes,{properties:"discard", mergeRels:true}) yield node return *
		`,
		map[string]interface{}{
			"localFullName":  fmt.Sprintf("%s|%s|%s", pkgName, moduleName, localName),
			"exportFullName": fmt.Sprintf("%s|%s|%s", pkgName, moduleName, exportName),
			"exportName":     exportName,
			"packageName":    pkgName,
			"fullModuleName": fmt.Sprintf("%s|%s", pkgName, moduleName),
			"moduleName":     moduleName,
		})

	if err != nil {
		return errors.Wrap(err, "error merging export and local function nodes")
	}

	return nil
}

func (e *ExportEdgeCreator) getLocalName(export resultprocessing.Export) string {
	if export.Local != "" {
		return export.Local
	}
	return export.Identifier
}
