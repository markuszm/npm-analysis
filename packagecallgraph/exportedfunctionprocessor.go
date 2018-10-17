package packagecallgraph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

type ExportedFunctionProcessor struct {
	inputFile     string
	outputFile    string
	logger        *zap.SugaredLogger
	mysqlDatabase *sql.DB
}

func NewExportedFunctionProcessor(inputFile, outputFile string, sql *sql.DB, logger *zap.SugaredLogger) *ExportedFunctionProcessor {
	return &ExportedFunctionProcessor{inputFile: inputFile, outputFile: outputFile, mysqlDatabase: sql, logger: logger}
}

func (e *ExportedFunctionProcessor) WriteProcessedExportResults() error {
	file, err := os.Open(e.inputFile)
	if err != nil {
		return errors.Wrap(err, "error opening export result file - does it exist?")
	}

	decoder := json.NewDecoder(file)

	exportedFunctionsSet := make(map[string]bool, 0)
	localFunctionToExportedFunctionMap := make(map[string]string, 0)

	progress := 0
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

		exports, err := resultprocessing.TransformToDynamicExports(result.Result)
		if err != nil {
			e.logger.With("package", result.Name).Error(err)
		}

		pkgName := result.Name

		for _, export := range exports {
			moduleNames := export.Locations
			localName := e.getLocalName(export)

			mainModule := getMainModuleForPackage(e.mysqlDatabase, pkgName)

			exportFullName := fmt.Sprintf("%s|%s|%s", pkgName, mainModule, export.Name)

			// if no locations are found we assume that the method is an redirect export or we just did not found it and keep it on the main module as actual export
			if len(moduleNames) == 0 {
				exportedFunctionsSet[exportFullName] = true
			}

			// if locations are found we can merge the local function node (if it exists) with the exported function node
			for _, module := range moduleNames {
				moduleName := trimExt(module.File)

				localFullName := fmt.Sprintf("%s|%s|%s", pkgName, moduleName, localName)

				exportedFunctionsSet[exportFullName] = true

				if localFullName == exportFullName {
					continue
				}
				localFunctionToExportedFunctionMap[localFullName] = exportFullName
			}
		}

		if progress%1000 == 0 {
			e.logger.Infof("Finished %v number of packages", progress)
		}

		progress++
	}

	exportResults := ExportResults{
		ExportedFunctions: exportedFunctionsSet,
		LocalFunctionsMap: localFunctionToExportedFunctionMap,
	}

	bytes, err := json.Marshal(exportResults)
	if err != nil {
		return errors.Wrap(err, "cannot marshal results")
	}

	err = ioutil.WriteFile(e.outputFile, bytes, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "cannot write results to output file")
	}

	return nil
}

func (e *ExportedFunctionProcessor) getLocalName(export resultprocessing.DynamicExport) string {
	if export.InternalName != "" {
		return export.InternalName
	}
	return export.Name
}

type ExportResults struct {
	ExportedFunctions map[string]bool
	LocalFunctionsMap map[string]string
}
