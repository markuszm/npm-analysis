package packagecallgraph

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"sync"
)

type CallEdgeCreator struct {
	neo4jUrl  string
	inputFile string
	logger    *zap.SugaredLogger
	workers   int
}

func NewCallEdgeCreator(neo4jUrl, callgraphInput string, workerNumber int, logger *zap.SugaredLogger) *CallEdgeCreator {
	return &CallEdgeCreator{neo4jUrl: neo4jUrl, inputFile: callgraphInput, workers: workerNumber, logger: logger}
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

func (g *CallEdgeCreator) insertCallIntoGraph(pkgName string, call resultprocessing.Call, database graph.Database) error {
	// TODO: check if imported name exists for functionName - needs imports somewhere stored

	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)
	_, err := database.Exec(`
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (l:LocalFunction {name: {fullLocalFunctionName}, functionName: {fromFunction}})
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
				MERGE (m1:Module {name: {fullModuleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}})
				MERGE (called:%s {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`, getFunctionType(call)),
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", pkgName, m),
					"requiredModuleName":     m,
					"fromFunctionName":       fromFunctionFullName,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting required module %s for call %s in package %s", m, call, pkgName)
			}
		} else {
			_, err = database.Exec(fmt.Sprintf(`
				MERGE (m1:Module {name: {fullModuleName}})
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}})
				MERGE (called:%s {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`, getFunctionType(call)),
				map[string]interface{}{
					// TODO: replace main magic string with real main module name
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", m, "main"),
					"requiredModuleName":     "main",
					"requiredPackageName":    m,
					"packageName":            pkgName,
					"fromFunctionName":       fromFunctionFullName,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", m, "main", call.ToFunction),
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
				MERGE (m1:Module {name: {fullModuleName}})
				MERGE (from:LocalFunction {name: {fromFunctionName}})
				MERGE (called:LocalFunction {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (from)-[:CALL]->(called)
				MERGE (m1)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"fromFunctionName":       fromFunctionFullName,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting localfunction call %s in package %s", call, pkgName)
			}
		} else {
			_, err = database.Exec(`
				MERGE (m1:Module {name: {fullModuleName}})
				MERGE (m2:Module {name: {receiver}})
				MERGE (from:LocalFunction {name: {fromFunctionName}})
				MERGE (called:ExportedFunction {name: {fullCalledFunctionName}, functionName: {calledFunctionName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"receiver":               call.Receiver,
					"fromFunctionName":       fromFunctionFullName,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s", call.Receiver, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				})
			if err != nil {
				return errors.Wrapf(err, "error inserting call %s in package %s", call, pkgName)
			}
		}
	}

	return nil
}

func getFunctionType(call resultprocessing.Call) string {
	if call.IsLocal || call.Receiver == "this" {
		return "LocalFunction"
	}
	return "ExportedFunction"
}
