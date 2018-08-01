package codeanalysis

import (
	"bytes"
	"encoding/json"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type UsedDependenciesAnalysis struct {
	logger             *zap.SugaredLogger
	importAnalysisPath string
}

func NewUsedDependenciesAnalysis(logger *zap.SugaredLogger, importAnalysisPath string) *UsedDependenciesAnalysis {
	return &UsedDependenciesAnalysis{logger, importAnalysisPath}
}

func (d *UsedDependenciesAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	importAnalysisResult, err := ExecuteCommand(d.importAnalysisPath, packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "error executing ast analysis")
	}

	var analysisResult []resultprocessing.Import
	err = json.Unmarshal([]byte(importAnalysisResult), &analysisResult)
	if err != nil {
		d.logger.Error(importAnalysisResult)
		return nil, errors.Wrap(err, "error parsing result")
	}

	packageMetadata, err := d.parsePackageJSON(packagePath)

	if err != nil {
		return nil, errors.Wrap(err, "error parsing package json")
	}

	requiredSet := make(map[string]bool, 0)
	importedSet := make(map[string]bool, 0)
	dependencySet := make(map[string]bool, 0)

	used := make([]string, 0)

	for d, _ := range packageMetadata.Dependencies {
		dependencySet[d] = true
	}
	// leaves out dev dependencies -> only some packages have tests included (bad practice)
	//for d, _ := range packageMetadata.DevDependencies {
	//	dependencySet[d] = true
	//}
	for d, _ := range packageMetadata.PeerDependencies {
		dependencySet[d] = true
	}
	for d, _ := range packageMetadata.OptionalDependencies {
		dependencySet[d] = true
	}

	for _, i := range analysisResult {
		module := i.ModuleName
		if module != "" && !IsLocalImport(module) && !requiredSet[module] && !importedSet[module] {
			if i.BundleType == "es6" {
				importedSet[module] = true
			} else {
				requiredSet[module] = true
			}

			if dependencySet[module] || isSubmodule(dependencySet, module) {
				used = append(used, module)
			}
		}
	}

	required := make([]string, len(requiredSet))
	i := 0
	for k := range requiredSet {
		required[i] = k
		i++
	}

	imported := make([]string, len(importedSet))
	i = 0
	for k := range importedSet {
		imported[i] = k
		i++
	}

	dependencies := make([]string, len(dependencySet))
	i = 0
	for k := range dependencySet {
		dependencies[i] = k
		i++
	}

	result := DependencyResult{
		Required:     required,
		Imported:     imported,
		Dependencies: dependencies,
		Used:         used,
	}

	return result, nil
}
func isSubmodule(dependencies map[string]bool, module string) bool {
	for d, ok := range dependencies {
		if ok {
			if strings.HasPrefix(module, d) {
				return true
			}
		}
	}
	return false
}

func IsLocalImport(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

func (d *UsedDependenciesAnalysis) parsePackageJSON(packagePath string) (MinimalPackage, error) {
	var packageMetadata MinimalPackage

	pathToPackageJSONStandard := path.Join(packagePath, "package", "package.json")
	pathToPackageJSONFallback := path.Join(packagePath, "package.json")

	pathToPackageJSON := pathToPackageJSONStandard
	_, err := os.Stat(pathToPackageJSONStandard)
	if err != nil {
		_, err := os.Stat(pathToPackageJSONFallback)

		if err == nil {
			pathToPackageJSON = pathToPackageJSONFallback
		} else {
			// fallback to walk through folder and take first package.json -> in worst case this could be the wrong one

			err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return errors.Wrap(err, "error finding package json")
				}
				if info.Name() == "package.json" {
					pathToPackageJSON = path
				}
				return nil
			})
			if err != nil {
				d.logger.Error(errors.Wrap(err, "error finding package json").Error())
				return packageMetadata, nil
			}
		}
	}

	fileContents, err := ioutil.ReadFile(pathToPackageJSON)
	if err != nil {
		d.logger.Errorf(errors.Wrap(err, "error reading package json").Error())
		return packageMetadata, nil
	}

	// remove BOM prefix see https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	fileContents = bytes.TrimPrefix(fileContents, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(fileContents, &packageMetadata)
	if err != nil {
		d.logger.Debug(string(fileContents))
		d.logger.Errorf(errors.Wrap(err, "error parsing package json").Error())
		return packageMetadata, nil
	}

	return packageMetadata, err
}

type DependencyResult struct {
	Required     []string
	Imported     []string
	Dependencies []string
	Used         []string
}

type MinimalPackage struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}
