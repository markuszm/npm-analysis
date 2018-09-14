package packagecallgraph

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type CallEdgeCreatorCSV struct {
	mysqlDatabase *sql.DB
	inputFile     string
	logger        *zap.SugaredLogger
	workers       int
	outputFolder  string
}

func NewCallEdgeCreatorCSV(output, callgraphInput string, workerNumber int, sql *sql.DB, logger *zap.SugaredLogger) *CallEdgeCreatorCSV {
	return &CallEdgeCreatorCSV{inputFile: callgraphInput, workers: workerNumber, logger: logger, mysqlDatabase: sql, outputFolder: output}
}

func (c *CallEdgeCreatorCSV) Exec() error {
	err := CreateHeaderFiles(c.outputFolder)
	if err != nil {
		c.logger.Fatalw("could not create csv header files", "err", err)
	}

	file, err := os.Open(c.inputFile)
	if err != nil {
		return errors.Wrap(err, "error opening callgraph result file - does it exist?")
	}

	decoder := json.NewDecoder(file)

	packageWorkerWait := sync.WaitGroup{}
	csvWorkerWait := sync.WaitGroup{}
	jobs := make(chan model.PackageResult, c.workers)

	csvChannels := CSVChannels{
		PackageChan:  make(chan WriteObject, 10),
		ModuleChan:   make(chan WriteObject, 10),
		ClassChan:    make(chan WriteObject, 10),
		FunctionChan: make(chan WriteObject, 10),
		RelationChan: make(chan WriteObject, 10),
	}

	csvWorkerWait.Add(5)
	go c.csvWorker(path.Join(c.outputFolder, "packages.csv"), csvChannels.PackageChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "modules.csv"), csvChannels.ModuleChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "classes.csv"), csvChannels.ClassChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "functions.csv"), csvChannels.FunctionChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "relations.csv"), csvChannels.RelationChan, &csvWorkerWait)

	for w := 1; w <= c.workers; w++ {
		packageWorkerWait.Add(1)
		go c.queryWorker(w, jobs, csvChannels, &packageWorkerWait)
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

	close(csvChannels.PackageChan)
	close(csvChannels.ModuleChan)
	close(csvChannels.ClassChan)
	close(csvChannels.FunctionChan)
	close(csvChannels.RelationChan)
	csvWorkerWait.Wait()

	return err
}

func (c *CallEdgeCreatorCSV) csvWorker(filePath string, packages chan WriteObject, workerWait *sync.WaitGroup) {
	file, err := os.Create(filePath)

	if err != nil {
		c.logger.Fatal("Cannot create result file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	i := 0

	for p := range packages {
		var fields []string
		fields = append(fields, p.GetFields()...)
		err = writer.Write(fields)
		if err != nil {
			c.logger.Fatalw("Cannot write to result file", "err", err)
		}
		if i%1000 == 0 {
			writer.Flush()
		}
		i++
	}

	workerWait.Done()
}

func (c *CallEdgeCreatorCSV) queryWorker(workerId int, jobs chan model.PackageResult, csvChannels CSVChannels, workerWait *sync.WaitGroup) {
	for j := range jobs {
		pkg := j.Name

		calls, err := resultprocessing.TransformToCalls(j.Result)
		if err != nil {
			c.logger.With("package", j.Name).Error(err)
		}

		receiverModuleMap := make(map[string][]string, 0)

		for _, call := range calls {
			if len(call.Modules) > 0 {
				receiverModuleMap[call.FromModule+call.Receiver] = call.Modules
			}

			c.createCSVRows(pkg, call, receiverModuleMap, csvChannels)
		}

		c.logger.Debugf("Worker: %v, Package: %s, Calls %v", workerId, j.Name, len(calls))

		// cleanup allocated slices by assigning nil
		calls = nil
		receiverModuleMap = nil
	}
	workerWait.Done()
}

func (c *CallEdgeCreatorCSV) createCSVRows(pkgName string, call resultprocessing.Call, receiverModuleMap map[string][]string, csvChannels CSVChannels) {
	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)
	fullModuleName := fmt.Sprintf("%s|%s", pkgName, call.FromModule)

	csvChannels.ModuleChan <- &Module{name: fullModuleName, moduleName: call.FromModule}
	csvChannels.PackageChan <- &Package{name: pkgName}
	csvChannels.RelationChan <- &Relation{startID: pkgName, endID: fullModuleName, relType: containsModule}
	csvChannels.FunctionChan <- &Function{
		name:         fromFunctionFullName,
		functionName: call.FromFunction,
		functionType: "local",
	}
	csvChannels.RelationChan <- &Relation{
		startID: fullModuleName,
		endID:   fromFunctionFullName,
		relType: containsFunction,
	}

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
			fullRequiredModuleName := fmt.Sprintf("%s|%s", pkgName, m)
			csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
			csvChannels.RelationChan <- &Relation{
				startID: fullModuleName,
				endID:   fullRequiredModuleName,
				relType: requiresModule,
			}

			calledFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction)
			csvChannels.FunctionChan <- &Function{
				name:         calledFunctionFullName,
				functionName: call.ToFunction,
				functionType: getFunctionType(call),
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   calledFunctionFullName,
				relType: callRelation,
			}

			csvChannels.RelationChan <- &Relation{
				startID: fullRequiredModuleName,
				endID:   calledFunctionFullName,
				relType: containsFunction,
			}
		} else if call.ClassName != "" {
			// case where call is to outside module class function
			fullRequiredModuleName := fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName)
			csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
			csvChannels.RelationChan <- &Relation{
				startID: fullModuleName,
				endID:   fullRequiredModuleName,
				relType: requiresModule,
			}

			csvChannels.PackageChan <- &Package{name: requiredPackageName}
			csvChannels.RelationChan <- &Relation{
				startID: pkgName,
				endID:   requiredPackageName,
				relType: requiresPackage,
			}
			csvChannels.RelationChan <- &Relation{
				startID: requiredPackageName,
				endID:   fullRequiredModuleName,
				relType: containsModule,
			}

			classFunctionFullName := fmt.Sprintf("%s|%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName, call.ToFunction)
			csvChannels.FunctionChan <- &Function{
				name:         classFunctionFullName,
				functionName: call.ToFunction,
				functionType: "class",
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   classFunctionFullName,
				relType: callRelation,
			}

			classFullName := fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName)
			csvChannels.ClassChan <- &Class{name: classFullName, className: call.ClassName}
			csvChannels.RelationChan <- &Relation{
				startID: fullRequiredModuleName,
				endID:   classFullName,
				relType: containsClass,
			}
			csvChannels.RelationChan <- &Relation{
				startID: classFullName,
				endID:   classFunctionFullName,
				relType: containsClassFunction,
			}

		} else {
			// case where call is to outside module
			fullRequiredModuleName := fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName)
			csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
			csvChannels.RelationChan <- &Relation{
				startID: fullModuleName,
				endID:   fullRequiredModuleName,
				relType: requiresModule,
			}

			csvChannels.PackageChan <- &Package{name: requiredPackageName}
			csvChannels.RelationChan <- &Relation{
				startID: pkgName,
				endID:   requiredPackageName,
				relType: requiresPackage,
			}
			csvChannels.RelationChan <- &Relation{
				startID: requiredPackageName,
				endID:   fullRequiredModuleName,
				relType: containsModule,
			}

			calledFunctionFullName := fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ToFunction)

			csvChannels.FunctionChan <- &Function{
				name:         calledFunctionFullName,
				functionName: call.ToFunction,
				functionType: getFunctionType(call),
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   calledFunctionFullName,
				relType: callRelation,
			}

			csvChannels.RelationChan <- &Relation{
				startID: fullRequiredModuleName,
				endID:   calledFunctionFullName,
				relType: containsFunction,
			}
		}

	}

	// special case where modules is empty
	if len(modules) == 0 {
		if call.IsLocal || call.Receiver == "this" {
			calledFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction)

			csvChannels.FunctionChan <- &Function{
				name:         calledFunctionFullName,
				functionName: call.ToFunction,
				functionType: "local",
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   calledFunctionFullName,
				relType: callRelation,
			}

			csvChannels.RelationChan <- &Relation{
				startID: fullModuleName,
				endID:   calledFunctionFullName,
				relType: containsFunction,
			}
		} else if call.ClassName != "" {
			classFunctionFullName := fmt.Sprintf("%s|%s", call.ClassName, call.ToFunction)

			csvChannels.FunctionChan <- &Function{
				name:         classFunctionFullName,
				functionName: call.ToFunction,
				functionType: "class",
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   classFunctionFullName,
				relType: callRelation,
			}

			csvChannels.ClassChan <- &Class{name: call.ClassName, className: call.ClassName}
			csvChannels.RelationChan <- &Relation{
				startID: call.ClassName,
				endID:   classFunctionFullName,
				relType: containsClassFunction,
			}

		} else {
			csvChannels.FunctionChan <- &Function{
				name:         call.ToFunction,
				functionName: call.ToFunction,
				functionType: "export",
			}
			csvChannels.RelationChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   call.ToFunction,
				relType: callRelation,
			}
		}
	}
}

func (c *CallEdgeCreatorCSV) getModuleNameForPackageImport(moduleName string) string {
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
