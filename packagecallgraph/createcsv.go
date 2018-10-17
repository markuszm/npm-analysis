package packagecallgraph

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type CallEdgeCreatorCSV struct {
	mysqlDatabase  *sql.DB
	callgraphInput string
	exportsInput   string
	logger         *zap.SugaredLogger
	workers        int
	outputFolder   string
	exportResults  ExportResults
}

func NewCallEdgeCreatorCSV(output, callgraphInput string, exportsInput string, workerNumber int, sql *sql.DB, logger *zap.SugaredLogger) *CallEdgeCreatorCSV {
	return &CallEdgeCreatorCSV{callgraphInput: callgraphInput, exportsInput: exportsInput, workers: workerNumber, logger: logger, mysqlDatabase: sql, outputFolder: output}
}

func (c *CallEdgeCreatorCSV) Exec() error {
	err := CreateHeaderFiles(c.outputFolder)
	if err != nil {
		c.logger.Fatalw("could not create csv header files", "err", err)
	}

	file, err := os.Open(c.callgraphInput)
	if err != nil {
		return errors.Wrap(err, "error opening callgraph result file - does it exist?")
	}

	bytes, err := ioutil.ReadFile(c.exportsInput)
	if err != nil {
		return errors.Wrap(err, "error opening exports result file - does it exist?")
	}
	exportResults := ExportResults{}
	err = json.Unmarshal(bytes, exportResults)
	if err != nil {
		return errors.Wrap(err, "error in unmarshal of exports")
	}
	c.exportResults = exportResults

	decoder := json.NewDecoder(file)

	packageWorkerWait := sync.WaitGroup{}
	csvWorkerWait := sync.WaitGroup{}
	jobs := make(chan model.PackageResult, c.workers)

	csvChannels := CSVChannels{
		PackageChan:               make(chan WriteObject, 10),
		ModuleChan:                make(chan WriteObject, 10),
		ClassChan:                 make(chan WriteObject, 10),
		FunctionChan:              make(chan WriteObject, 10),
		CallsChan:                 make(chan WriteObject, 10),
		ContainsClassChan:         make(chan WriteObject, 10),
		ContainsClassFunctionChan: make(chan WriteObject, 10),
		ContainsFunctionChan:      make(chan WriteObject, 10),
		ContainsModuleChan:        make(chan WriteObject, 10),
		RequiresModuleChan:        make(chan WriteObject, 10),
		RequiresPackageChan:       make(chan WriteObject, 10),
	}

	csvWorkerWait.Add(11)
	go c.csvWorker(path.Join(c.outputFolder, "packages.csv"), csvChannels.PackageChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "modules.csv"), csvChannels.ModuleChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "classes.csv"), csvChannels.ClassChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "functions.csv"), csvChannels.FunctionChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "calls.csv"), csvChannels.CallsChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "containsclass.csv"), csvChannels.ContainsClassChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "containsclassfunction.csv"), csvChannels.ContainsClassFunctionChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "containsfunction.csv"), csvChannels.ContainsFunctionChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "containsmodule.csv"), csvChannels.ContainsModuleChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "requiresmodule.csv"), csvChannels.RequiresModuleChan, &csvWorkerWait)
	go c.csvWorker(path.Join(c.outputFolder, "requirespackage.csv"), csvChannels.RequiresPackageChan, &csvWorkerWait)

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
	close(csvChannels.CallsChan)
	close(csvChannels.ContainsClassChan)
	close(csvChannels.ContainsClassFunctionChan)
	close(csvChannels.ContainsFunctionChan)
	close(csvChannels.ContainsModuleChan)
	close(csvChannels.RequiresModuleChan)
	close(csvChannels.RequiresPackageChan)
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
			if hasValidModules(call) {
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
	if !isValidCall(call) {
		return
	}

	fromFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.FromFunction)

	// map local function name to exported function name if it exists
	fromFunctionFullName = c.mapToExportedFunction(fromFunctionFullName)

	fullModuleName := fmt.Sprintf("%s|%s", pkgName, call.FromModule)

	csvChannels.ModuleChan <- &Module{name: fullModuleName, moduleName: call.FromModule}
	csvChannels.PackageChan <- &Package{name: pkgName}
	csvChannels.ContainsModuleChan <- &Relation{startID: pkgName, endID: fullModuleName}
	csvChannels.FunctionChan <- &Function{
		name:         fromFunctionFullName,
		functionName: call.FromFunction,
		functionType: c.getFunctionTypeForFunctionName(fromFunctionFullName),
	}
	csvChannels.ContainsFunctionChan <- &Relation{
		startID: fullModuleName,
		endID:   fromFunctionFullName,
	}

	modules := call.Modules
	if len(modules) == 0 && call.Receiver != "" {
		refModules, exists := receiverModuleMap[call.FromModule+call.Receiver]
		if exists && len(refModules) != 0 {
			modules = refModules
		}
	}

	if hasValidModules(call) {
		for _, m := range modules {
			importedModuleName := getMainModuleForPackage(c.mysqlDatabase, m)
			requiredPackageName := getRequiredPackageName(m)

			if codeanalysis.IsLocalImport(m) {
				fullRequiredModuleName := fmt.Sprintf("%s|%s", pkgName, m)
				csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
				csvChannels.RequiresModuleChan <- &Relation{
					startID: fullModuleName,
					endID:   fullRequiredModuleName,
				}

				calledFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, m, call.ToFunction)
				calledFunctionFullName = c.mapToExportedFunction(calledFunctionFullName)

				functionType := c.getFunctionTypeForFunctionName(calledFunctionFullName)
				if functionType == "unknown" {
					functionType = c.getFunctionTypeFromCall(call)
				}

				csvChannels.FunctionChan <- &Function{
					name:         calledFunctionFullName,
					functionName: call.ToFunction,
					functionType: functionType,
				}
				csvChannels.CallsChan <- &Relation{
					startID: fromFunctionFullName,
					endID:   calledFunctionFullName,
				}

				csvChannels.ContainsFunctionChan <- &Relation{
					startID: fullRequiredModuleName,
					endID:   calledFunctionFullName,
				}
			} else if call.ClassName != "" {
				// case where call is to outside module class function
				fullRequiredModuleName := fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName)
				csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
				csvChannels.RequiresModuleChan <- &Relation{
					startID: fullModuleName,
					endID:   fullRequiredModuleName,
				}

				csvChannels.PackageChan <- &Package{name: requiredPackageName}
				csvChannels.RequiresPackageChan <- &Relation{
					startID: pkgName,
					endID:   requiredPackageName,
				}
				csvChannels.ContainsModuleChan <- &Relation{
					startID: requiredPackageName,
					endID:   fullRequiredModuleName,
				}

				classFunctionFullName := fmt.Sprintf("%s|%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName, call.ToFunction)
				csvChannels.FunctionChan <- &Function{
					name:         classFunctionFullName,
					functionName: call.ToFunction,
					functionType: "class",
				}
				csvChannels.CallsChan <- &Relation{
					startID: fromFunctionFullName,
					endID:   classFunctionFullName,
				}

				classFullName := fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ClassName)
				csvChannels.ClassChan <- &Class{name: classFullName, className: call.ClassName}
				csvChannels.ContainsClassChan <- &Relation{
					startID: fullRequiredModuleName,
					endID:   classFullName,
				}
				csvChannels.ContainsClassFunctionChan <- &Relation{
					startID: classFullName,
					endID:   classFunctionFullName,
				}

			} else {
				// case where call is to outside module
				fullRequiredModuleName := fmt.Sprintf("%s|%s", requiredPackageName, importedModuleName)
				csvChannels.ModuleChan <- &Module{name: fullRequiredModuleName, moduleName: m}
				csvChannels.RequiresModuleChan <- &Relation{
					startID: fullModuleName,
					endID:   fullRequiredModuleName,
				}

				csvChannels.PackageChan <- &Package{name: requiredPackageName}
				csvChannels.RequiresPackageChan <- &Relation{
					startID: pkgName,
					endID:   requiredPackageName,
				}
				csvChannels.ContainsModuleChan <- &Relation{
					startID: requiredPackageName,
					endID:   fullRequiredModuleName,
				}

				calledFunctionFullName := fmt.Sprintf("%s|%s|%s", requiredPackageName, importedModuleName, call.ToFunction)
				calledFunctionFullName = c.mapToExportedFunction(calledFunctionFullName)

				functionType := c.getFunctionTypeForFunctionName(calledFunctionFullName)
				if functionType == "unknown" {
					functionType = c.getFunctionTypeFromCall(call)
				}

				csvChannels.FunctionChan <- &Function{
					name:         calledFunctionFullName,
					functionName: call.ToFunction,
					functionType: functionType,
				}
				csvChannels.CallsChan <- &Relation{
					startID: fromFunctionFullName,
					endID:   calledFunctionFullName,
				}

				csvChannels.ContainsFunctionChan <- &Relation{
					startID: fullRequiredModuleName,
					endID:   calledFunctionFullName,
				}
			}

		}
	} else {
		// special case where modules is empty
		if call.IsLocal || call.Receiver == "this" {
			calledFunctionFullName := fmt.Sprintf("%s|%s|%s", pkgName, call.FromModule, call.ToFunction)
			calledFunctionFullName = c.mapToExportedFunction(calledFunctionFullName)

			csvChannels.FunctionChan <- &Function{
				name:         calledFunctionFullName,
				functionName: call.ToFunction,
				functionType: c.getFunctionTypeForFunctionName(calledFunctionFullName),
			}
			csvChannels.CallsChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   calledFunctionFullName,
			}

			csvChannels.ContainsFunctionChan <- &Relation{
				startID: fullModuleName,
				endID:   calledFunctionFullName,
			}
		} else if call.ClassName != "" {
			classFunctionFullName := fmt.Sprintf("%s|%s", call.ClassName, call.ToFunction)

			csvChannels.FunctionChan <- &Function{
				name:         classFunctionFullName,
				functionName: call.ToFunction,
				functionType: "class",
			}
			csvChannels.CallsChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   classFunctionFullName,
			}

			csvChannels.ClassChan <- &Class{name: call.ClassName, className: call.ClassName}
			csvChannels.ContainsClassFunctionChan <- &Relation{
				startID: call.ClassName,
				endID:   classFunctionFullName,
			}

		} else {
			csvChannels.FunctionChan <- &Function{
				name:         call.ToFunction,
				functionName: call.ToFunction,
				functionType: "export",
			}
			csvChannels.CallsChan <- &Relation{
				startID: fromFunctionFullName,
				endID:   call.ToFunction,
			}
		}
	}
}

func (c *CallEdgeCreatorCSV) getFunctionTypeForFunctionName(functionName string) string {
	if c.exportResults.ExportedFunctions[functionName] {
		return "actualExport"
	}
	return "unknown"
}

func (c *CallEdgeCreatorCSV) mapToExportedFunction(functionName string) string {
	actualExportedFunctionName := c.exportResults.LocalFunctionsMap[functionName]
	if actualExportedFunctionName != "" {
		functionName = actualExportedFunctionName
	}
	return functionName
}

func (c *CallEdgeCreatorCSV) getFunctionTypeFromCall(call resultprocessing.Call) string {
	if call.IsLocal || call.Receiver == "this" {
		return "local"
	}
	return "export"
}
