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
	"time"
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

	packageWorkerWait := sync.WaitGroup{}
	queryWorkerWait := sync.WaitGroup{}
	jobs := make(chan model.PackageResult, c.workers)
	queries := make(chan Neo4jQuery, c.workers)

	for w := 1; w <= c.workers; w++ {
		packageWorkerWait.Add(1)
		go c.queryWorker(w, jobs, queries, &packageWorkerWait)
	}

	for w := 1; w <= c.workers; w++ {
		queryWorkerWait.Add(1)
		go c.execWorker(w, queries, &queryWorkerWait)
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
	packageWorkerWait.Wait()
	close(queries)
	queryWorkerWait.Wait()

	return err
}

func (c *CallEdgeCreator) queryWorker(workerId int, jobs chan model.PackageResult, queryChannel chan Neo4jQuery, workerWait *sync.WaitGroup) {
	for j := range jobs {
		pkg := j.Name

		calls, err := resultprocessing.TransformToCalls(j.Result)
		if err != nil {
			c.logger.With("package", j.Name).Error(err)
		}

		receiverModuleMap := make(map[string][]string, 0)

		for _, call := range calls {
			if hasValidModules(call) {
				receiverModuleMap[call.FromModule+call.Receiver] = call.Modules
			}

			c.createQueries(pkg, call, receiverModuleMap, queryChannel)
		}

		c.logger.Debugf("Worker: %v, Package: %s, Calls %v", workerId, j.Name, len(calls))

		// cleanup allocated slices by assigning nil
		calls = nil
		receiverModuleMap = nil
	}
	workerWait.Done()
}

func (c *CallEdgeCreator) createQueries(pkgName string, call resultprocessing.Call, receiverModuleMap map[string][]string, queryChannel chan Neo4jQuery) {
	if !isValidCall(call) {
		return
	}

	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)
	fullModuleName := fmt.Sprintf("%s|%s", pkgName, call.FromModule)

	queryChannel <- Neo4jQuery{`
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (p:Package {name: {packageName}})
		MERGE (p)-[:CONTAINS_MODULE]->(m)
`,
		map[string]interface{}{
			"fullModuleName": fullModuleName,
			"moduleName":     call.FromModule,
			"packageName":    pkgName,
		}}

	queryChannel <- Neo4jQuery{`
		MERGE (m:Module {name: {fullModuleName}, moduleName: {moduleName}})
		MERGE (f:Function {name: {fullLocalFunctionName}}) ON CREATE SET f.functionName = {fromFunction}, f.functionType = "local"
		MERGE (m)-[:CONTAINS_FUNCTION]->(f)`,
		map[string]interface{}{
			"fullModuleName":        fullModuleName,
			"moduleName":            call.FromModule,
			"fullLocalFunctionName": fromFunctionFullName,
			"fromFunction":          call.FromFunction,
		}}

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
			queryChannel <- Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				`,
				map[string]interface{}{
					"fullModuleName":         fullModuleName,
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", pkgName, m),
					"requiredModuleName":     m,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", pkgName, m),
					"requiredModuleName":     m,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}}

		} else if call.ClassName != "" {
			// case where call is to outside module class function
			queryChannel <- Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				`,
				map[string]interface{}{
					"fullModuleName":         fullModuleName,
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				`,
				map[string]interface{}{
					"packageName":         pkgName,
					"requiredPackageName": requiredPackageName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				`,
				map[string]interface{}{
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"requiredPackageName":    requiredPackageName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":  fromFunctionFullName,
					"fromFunction":      call.FromFunction,
					"fullClassFunction": fmt.Sprintf("%s|%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName, call.ToFunction),
					"classFunction":     call.ToFunction,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (c:Class {name: {fullClassName}, className: {className}})
				MERGE (m2)-[:CONTAINS_CLASS]->(c)
				`,
				map[string]interface{}{
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"fullClassName":          fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName),
					"className":              call.ClassName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (c:Class {name: {fullClassName}, className: {className}})
				MERGE (c)-[:CONTAINS_CLASS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullClassName":     fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName),
					"className":         call.ClassName,
					"fullClassFunction": fmt.Sprintf("%s|%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName, call.ToFunction),
					"classFunction":     call.ToFunction,
				}}

		} else {
			// case where call is to outside module
			queryChannel <- Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})	
				MERGE (m1)-[:REQUIRES_MODULE]->(m2)
				`,
				map[string]interface{}{
					"fullModuleName":         fullModuleName,
					"moduleName":             call.FromModule,
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (p1:Package {name: {packageName}})
				MERGE (p2:Package {name: {requiredPackageName}})
				MERGE (p1)-[:REQUIRES_PACKAGE]->(p2)
				`,
				map[string]interface{}{
					"requiredPackageName": requiredPackageName,
					"packageName":         pkgName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (p2:Package {name: {requiredPackageName}})	
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (p2)-[:CONTAINS_MODULE]->(m2)
				`,
				map[string]interface{}{
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"requiredPackageName":    requiredPackageName,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (m2:Module {name: {fullRequiredModuleName}, moduleName: {requiredModuleName}})
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = {calledFunctionType}
				MERGE (m2)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullRequiredModuleName": fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName),
					"requiredModuleName":     importedModuleName,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
					"calledFunctionType":     getFunctionType(call),
				}}
		}

	}

	// special case where modules is empty
	if len(modules) == 0 {
		if call.IsLocal || call.Receiver == "this" {
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = "local"
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":       fromFunctionFullName,
					"fromFunction":           call.FromFunction,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (m1:Module {name: {fullModuleName}, moduleName: {moduleName}})
				MERGE (called:Function {name: {fullCalledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = "local"	
				MERGE (m1)-[:CONTAINS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"fullModuleName":         fullModuleName,
					"moduleName":             call.FromModule,
					"fullCalledFunctionName": fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction),
					"calledFunctionName":     call.ToFunction,
				}}
		} else if call.ClassName != "" {
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":  fromFunctionFullName,
					"fromFunction":      call.FromFunction,
					"fullClassFunction": fmt.Sprintf("%s|%s", call.ClassName, call.ToFunction),
					"classFunction":     call.ToFunction,
				}}
			queryChannel <- Neo4jQuery{`
				MERGE (called:Function {name: {fullClassFunction}}) ON CREATE SET called.functionName = {classFunction}, called.functionType = "class"
				MERGE (c:Class {name: {className}, className: {className}})
				MERGE (c)-[:CONTAINS_CLASS_FUNCTION]->(called)
				`,
				map[string]interface{}{
					"className":         call.ClassName,
					"fullClassFunction": fmt.Sprintf("%s|%s", call.ClassName, call.ToFunction),
					"classFunction":     call.ToFunction,
				}}
		} else {
			queryChannel <- Neo4jQuery{`
				MERGE (from:Function {name: {fromFunctionName}}) ON CREATE SET from.functionName = {fromFunction}, from.functionType = "local"
				MERGE (called:Function {name: {calledFunctionName}}) ON CREATE SET called.functionName = {calledFunctionName}, called.functionType = "export"
				MERGE (from)-[:CALL]->(called)
				`,
				map[string]interface{}{
					"fromFunctionName":   fromFunctionFullName,
					"fromFunction":       call.FromFunction,
					"calledFunctionName": call.ToFunction,
				}}
		}
	}
}

func (c *CallEdgeCreator) execWorker(workerId int, jobs chan Neo4jQuery, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	err := neo4JDatabase.InitDB(c.neo4jUrl)
	if err != nil {
		c.logger.Fatal(err)
	}
	defer neo4JDatabase.Close()

	for q := range jobs {
		retry := 0
		_, err = neo4JDatabase.Exec(q.queryString, q.parameters)
		for retry < c.workers && err != nil {
			_, err = neo4JDatabase.Exec(q.queryString, q.parameters)
			time.Sleep(3 * time.Second)
			retry++
		}
		if err != nil {
			c.logger.With("error", err).Fatal("error inserting calls")
			continue
		}
	}
	workerWait.Done()
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

func isValidCall(call resultprocessing.Call) bool {
	if call.ToFunction == "" {
		return false
	}
	return true
}

func hasValidModules(call resultprocessing.Call) bool {
	for _, m := range call.Modules {
		if m == "" {
			return false
		}
	}
	return len(call.Modules) > 0
}
