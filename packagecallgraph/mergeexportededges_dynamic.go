package packagecallgraph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path"
	"strings"
	"sync"
)

type DynamicExportEdgeCreator struct {
	neo4jUrl      string
	mysqlDatabase *sql.DB
	inputFile     string
	logger        *zap.SugaredLogger
	workers       int
}

func NewDynamicExportEdgeCreator(neo4jUrl, inputFile string, workerNumber int, sql *sql.DB, logger *zap.SugaredLogger) *DynamicExportEdgeCreator {
	return &DynamicExportEdgeCreator{neo4jUrl: neo4jUrl, inputFile: inputFile, mysqlDatabase: sql, workers: workerNumber, logger: logger}
}

func (e *DynamicExportEdgeCreator) Exec() error {
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

func (e *DynamicExportEdgeCreator) worker(workerId int, jobs chan model.PackageResult, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	err := neo4JDatabase.InitDB(e.neo4jUrl)
	if err != nil {
		e.logger.Fatal(err)
	}
	defer neo4JDatabase.Close()

	for j := range jobs {
		pkg := j.Name

		exports, err := resultprocessing.TransformToDynamicExports(j.Result)
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
					continue
				}
			}
		}

		e.logger.Debugf("Worker: %v, Package: %s, Exports %v", workerId, j.Name, len(exports))
	}
	workerWait.Done()
}

func (e *DynamicExportEdgeCreator) addExportEdges(pkgName string, export resultprocessing.DynamicExport, database graph.Database) error {
	exportName := export.Name
	moduleNames := export.Locations
	localName := e.getLocalName(export)

	mainModule := getMainModuleForPackage(e.mysqlDatabase, pkgName)

	// if no locations are found we assume that the method is an redirect export or we just did not found it and keep it on the main module as actual export
	if len(moduleNames) == 0 {
		_, err := database.Exec(`
		MERGE (e:Function {name: {exportFullName}}) 
			ON CREATE SET e.functionName = {exportName}, e.functionType = "export", e.actualExport = true
			ON MATCH SET e.actualExport = true
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(e)
		`,
			map[string]interface{}{
				"exportFullName": fmt.Sprintf("%s|%s|%s", pkgName, mainModule, exportName),
				"exportName":     exportName,
				"packageName":    pkgName,
				"fullModuleName": fmt.Sprintf("%s|%s", pkgName, mainModule),
				"moduleName":     mainModule,
			})

		if err != nil {
			return errors.Wrap(err, "error adding exported method")
		}
	}

	// if locations are found we can merge the local function node (if it exists) with the exported function node
	for _, module := range moduleNames {
		moduleName := trimExt(module.File)

		_, err := database.Exec(`
		MATCH (l:Function {name: {localFullName}})
		MERGE (e:Function {name: {exportFullName}}) 
			ON CREATE SET e.functionName = {exportName}, e.functionType = "export", e.actualExport = true
			ON MATCH SET e.actualExport = true
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(e)
		WITH [e,l] as nodes CALL apoc.refactor.mergeNodes(nodes,{properties:"override", mergeRels:true}) yield node return *
		`,
			map[string]interface{}{
				"localFullName":  fmt.Sprintf("%s|%s|%s", pkgName, moduleName, localName),
				"exportFullName": fmt.Sprintf("%s|%s|%s", pkgName, mainModule, exportName),
				"exportName":     exportName,
				"packageName":    pkgName,
				"fullModuleName": fmt.Sprintf("%s|%s", pkgName, mainModule),
				"moduleName":     mainModule,
			})

		if err != nil {
			return errors.Wrapf(err, "error merging export and local function nodes with module: %s export: %s local function: %s", moduleName, exportName, localName)
		}
	}

	return nil
}

func (e *DynamicExportEdgeCreator) getLocalName(export resultprocessing.DynamicExport) string {
	if export.InternalName != "" {
		return export.InternalName
	}
	return export.Name
}

func trimExt(moduleName string) string {
	return strings.Replace(moduleName, path.Ext(moduleName), "", -1)
}
