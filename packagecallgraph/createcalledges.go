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

type Neo4jQuery struct {
	queryString string
	parameters  map[string]interface{}
}

func NewCallEdgeCreator(neo4jUrl, callgraphInput string, workerNumber int, sql *sql.DB, logger *zap.SugaredLogger) *CallEdgeCreator {
	return &CallEdgeCreator{neo4jUrl: neo4jUrl, inputFile: callgraphInput, workers: workerNumber, logger: logger, mysqlDatabase: sql}
}

func (c *CallEdgeCreator) Exec() error {
	file, err := os.Open(c.inputFile)
	if err != nil {
		return errors.Wrap(err, "error opening callgraph result file - does it exist?")
	}

	decoder := json.NewDecoder(file)

	workerWait := sync.WaitGroup{}

	jobs := make(chan model.PackageResult, 100)

	for w := 1; w <= c.workers; w++ {
		workerWait.Add(1)
		go c.worker(w, jobs, &workerWait)
	}

	for {
		result := model.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				c.logger.Debug("finished decoding result json")
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

func (c *CallEdgeCreator) worker(workerId int, jobs chan model.PackageResult, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	err := neo4JDatabase.InitDB(c.neo4jUrl)
	if err != nil {
		c.logger.Fatal(err)
	}
	defer neo4JDatabase.Close()

	for j := range jobs {
		pkg := j.Name

		calls, err := resultprocessing.TransformToCalls(j.Result)
		if err != nil {
			c.logger.With("package", j.Name).Error(err)
		}

		receiverModuleMap := make(map[string][]string, 0)

		var allQueries []Neo4jQuery

		for _, call := range calls {
			if len(call.Modules) > 0 {
				receiverModuleMap[call.FromModule+call.Receiver] = call.Modules
			}

			allQueries = append(allQueries, c.createQueries(pkg, call, receiverModuleMap, neo4JDatabase)...)

		}

		queries := make([]string, len(allQueries))
		parameters := make([]map[string]interface{}, len(allQueries))

		for i, q := range allQueries {
			queries[i] = q.queryString
			parameters[i] = q.parameters
		}

		retry := 0

		_, err = neo4JDatabase.ExecPipeline(false, queries, parameters...)
		if err != nil {
			for retry < 3 && err != nil {
				_, err = neo4JDatabase.ExecPipeline(false, queries, parameters...)
				retry++
			}
			if err != nil {
				c.logger.With("package", pkg, "error", err).Error("error inserting calls")
			}
		}

		c.logger.Debugf("Worker: %v, Package: %s, Calls %v", workerId, j.Name, len(calls))
	}
	workerWait.Done()
}

func (c *CallEdgeCreator) createQueries(pkgName string, call resultprocessing.Call, receiverModuleMap map[string][]string, database graph.Database) []Neo4jQuery {
	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)
	var queries []Neo4jQuery
	queries = append(queries, Neo4jQuery{`
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (f:Function {name: {fullLocalFunctionName}}) ON CREATE SET f.functionName = {fromFunction}, f.functionType = "local"
		MERGE (p)-[:CONTAINS_MODULE]->(m)
		MERGE (m)-[:CONTAINS_FUNCTION]->(f)`,
		map[string]interface{}{
			"fullModuleName":        fmt.Sprintf("%s|%s", pkgName, call.FromModule),
			"moduleName":            call.FromModule,
			"packageName":           pkgName,
			"fullLocalFunctionName": fromFunctionFullName,
			"fromFunction":          call.FromFunction,
		}})

	modules := call.Modules
	if len(modules) == 0 && call.Receiver != "" {
		refModules, exists := receiverModuleMap[call.FromModule+call.Receiver]
		if exists {
			modules = refModules
		}
	}

	for _, m := range modules {
		importedModuleName := c.getModuleNameForPackageImport(m)
		requiredPackageName := getRequiredPackageName(m)
		if codeanalysis.IsLocalImport(m) {
			queries = append(queries, Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", pkgName, m),
					"requiredModuleName":     m,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}})
		} else if call.ClassName != "" {
			queries = append(queries, Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (c:Class {name: {fullClassName}, className: {className}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_CLASS]->(c)
				MERGE (c)-[:CONTAINS_CLASS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"requiredPackageName":    requiredPackageName,
					"packageName":            pkgName,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullClassName":          fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName),
					"className":              call.ClassName,
					"fullClassFunction":      fmt.Sprintf("%s|%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName, call.ToFunction),
					"classFunction":          call.ToFunction,
				}})

		} else {
			queries = append(queries, Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				MERGE (from)-[:CALL]->(called)
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"requiredPackageName":    requiredPackageName,
					"packageName":            pkgName,
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}})
		}

	}

	// special case where modules is empty
	if len(modules) == 0 {
		if call.IsLocal || call.Receiver == "this" {
			queries = append(queries, Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = "local"
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
				}})
		} else if call.ClassName != "" {
			queries = append(queries, Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (c:Class {name: {className}, className: {className}})
				MERGE (from)-[:CALL]->(called)
				MERGE (c)-[:CONTAINS_CLASS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":    fmt.Sprintf("%s|%s", pkgName, call.FromModule),
					"moduleName":        call.FromModule,
					"fromFunctionName":  fromFunctionFullName,
					"fromFunction":      call.FromFunction,
					"className":         call.ClassName,
					"fullClassFunction": fmt.Sprintf("%s|%s", call.ClassName, call.ToFunction),
					"classFunction":     call.ToFunction,
				}})
		} else {
			queries = append(queries, Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {calledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = "export"
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":   fromFunctionFullName,
					"fromFunction":       call.FromFunction,
					"calledFunctionName": call.ToFunction,
				}})
		}
	}

	return queries
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
		return "local"
	}
	return "export"
}
