package packagecallgraph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type CallEdgeCreator struct {
	neo4jUrl      string
	mysqlDatabase *sql.DB
	inputFile     string
	logger        *zap.SugaredLogger
	workers       int
}

func NewCallEdgeCreator(neo4jUrl, callgraphInput string, workerNumber int, sql *sql.DB, logger *zap.SugaredLogger) *CallEdgeCreator {
	return &CallEdgeCreator{neo4jUrl: neo4jUrl, inputFile: callgraphInput, workers: workerNumber, logger: logger, mysqlDatabase: sql}
}

func (g *CallEdgeCreator) Exec() error {
	file, err := os.Open(g.inputFile)
	if err != nil {
		return errors.Wrap(err, "error opening callgraph result file - does it exist in input folder?")
	}

	decoder := json.NewDecoder(file)

	workerWait := sync.WaitGroup{}

	jobs := make(chan model.PackageResult, 100)

	for w := 1; w <= g.workers; w++ {
		workerWait.Add(1)
		go g.worker(w, jobs, &workerWait)
	}

	for {
		result := model.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				g.logger.Debug("finished decoding result json")
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

func (g *CallEdgeCreator) worker(workerId int, jobs chan model.PackageResult, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	defer neo4JDatabase.Close()
	err := neo4JDatabase.InitDB(g.neo4jUrl)

	if err != nil {
		g.logger.Fatal(err)
	}

	for j := range jobs {
		pkg := j.Name
		_, err := neo4JDatabase.Exec(`MERGE (:Package {name: {pkgName}})`, map[string]interface{}{"pkgName": pkg})
		if err != nil {
			g.logger.Fatalw("error creating package node", "package", pkg, "error", err)
		}

		calls, err := resultprocessing.TransformToCalls(j.Result)
		if err != nil {
			g.logger.Fatal(err)
		}

		for _, call := range calls {
			err := g.insertCallIntoGraph(pkg, call, neo4JDatabase)
			if err != nil {
				g.logger.Fatalw("error inserting call", "call", call, "error", err)
			}
		}

		g.logger.Debugf("Worker: %v, Package: %s, Calls %v", workerId, j.Name, len(calls))
	}
	workerWait.Done()
}

func (c *CallEdgeCreator) insertCallIntoGraph(pkgName string, call resultprocessing.Call, database graph.Database) error {
	// TODO: check if imported name exists for functionName - needs imports somewhere stored

	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)
	_, err := database.Exec(`
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (l:LocalFunction {name: {fullLocalFunctionName}, functionName: {fromFunction}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(l)`,
		map[string]interface{}{
			"fullModuleName":        fmt.Sprintf("%s|%s", pkgName, call.FromModule),
			"moduleName":            call.FromModule,
			"packageName":           pkgName,
			"fullLocalFunctionName": fromFunctionFullName,
			"fromFunction":          call.FromFunction,
		})
	if err != nil {
		return errors.Wrap(err, "error inserting module node")
	}

	// TODO: store receiver module relation for calls to same module when modules empty

	for _, m := range call.Modules {
		if codeanalysis.IsLocalImport(m) {
			_, err = database.Exec(fmt.Sprintf(`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}, functionName: {fromFunction}})
				MERGE (called:%s {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`, getFunctionType(call)),
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", pkgName, m),
					"requiredModuleName":     m,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting required module %s for call %s in package %s", m, call, pkgName)
			}
		} else {
			_, err = database.Exec(fmt.Sprintf(`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}, functionName: {fromFunction}})
				MERGE (called:%s {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`, getFunctionType(call)),
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", getRequiredPackageName(m), c.getModuleNameForPackageImport(m)),
					"requiredModuleName":     c.getModuleNameForPackageImport(m),
					"requiredPackageName":    getRequiredPackageName(m),
					"packageName":            pkgName,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", m, c.getModuleNameForPackageImport(m), call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting required module %s for call %s in package %s", m, call, pkgName)
			}
		}

	}

	// special case where modules is empty
	if len(call.Modules) == 0 {
		if call.IsLocal || call.Receiver == "this" {
			_, err = database.Exec(`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}, functionName: {fromFunction}})
				MERGE (called:LocalFunction {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (from)-[:CALL]->(called)
				MERGE (m1)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting localfunction call %s in package %s", call, pkgName)
			}
		} else if call.Receiver != "" {
			_, err = database.Exec(`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {receiver}, moduleName: {receiver}})
				MERGE (from:LocalFunction {name: {fromFunctionName}, functionName: {fromFunction}})
				MERGE (called:ExportedFunction {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"receiver":               call.Receiver,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s", call.Receiver, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting call %s in package %s", call, pkgName)
			}
		} else {
			_, err = database.Exec(`
				MERGE (from:LocalFunction {name: {fromFunctionName}, functionName: {fromFunction}})
				MERGE (called:ExportedFunction {name: {calledFunctionName}, functionName: {calledFunctionName}})
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":   fromFunctionFullName,
					"fromFunction":       call.FromFunction,
					"calledFunctionName": call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting call %s in package %s", call, pkgName)
			}
		}
	}

	return nil
}

func (c *CallEdgeCreator) getModuleNameForPackageImport(moduleName string) string {
	packageName := getRequiredPackageName(moduleName)
	if strings.Contains(moduleName, "/") && packageName != moduleName {
		moduleName := strings.Replace(moduleName, packageName+"/", "", -1)
		return moduleName
	}
	mainFile, err := database.MainFileForPackage(c.mysqlDatabase, packageName)
	if err != nil {
		c.logger.Fatalf("error getting mainFile from database for moduleName %s with error %s", moduleName, err)
	}
	// cleanup main file
	mainFile = strings.TrimSuffix(mainFile, filepath.Ext(mainFile))
	mainFile = strings.TrimLeft(mainFile, "./")

	if mainFile == "" {
		return "index"
	}
	return mainFile
}

func getRequiredPackageName(moduleName string) string {
	if strings.Contains(moduleName, "/") {
		parts := strings.Split(moduleName, "/")
		if strings.HasPrefix(moduleName, "@") {
			return fmt.Sprintf("%s/%s", parts[0], parts[1])
		}
		return parts[0]
	}
	return moduleName
}

func getFunctionType(call resultprocessing.Call) string {
	if call.IsLocal || call.Receiver == "this" {
		return "LocalFunction"
	}
	return "ExportedFunction"
}
