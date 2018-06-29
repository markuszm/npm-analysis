package codeanalysis

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type UsedDependenciesAnalysis struct {
}

func (d *UsedDependenciesAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	requiredResult, err := ExecuteCommand("grep", "-ohrP", "--include=*.js", "--include=*.ts", "--exclude-dir=node_modules", `(?<=require\().+?(?=\))`, packagePath)
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 1") {
			return nil, errors.Wrap(err, "error retrieving required packages")
		}
	}

	importResult, err := ExecuteCommand("grep", "-ohrP", "--include=*.js", "--include=*.ts", "--exclude-dir=node_modules", `^import .*(("|').+("|'));?`, packagePath)
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 1") {
			return nil, errors.Wrap(err, "error retrieving imported packages")
		}
	}

	packageMetadata, err := parsePackageJSON(packagePath)

	if err != nil {
		return nil, errors.Wrap(err, "error parsing package json")
	}

	requiredPackages := strings.Split(requiredResult, "\n")
	importedPackages := strings.Split(importResult, "\n")

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

	for _, r := range requiredPackages {
		module := stripQuotes(r)
		if module != "" && !IsLocalImport(module) && !requiredSet[module] {
			requiredSet[module] = true

			if dependencySet[module] {
				used = append(used, module)
			}
		}
	}

	for _, i := range importedPackages {
		if i == "" {
			continue
		}
		module := parseModuleFromImportStmt(i)

		if module != "" && !IsLocalImport(module) && !importedSet[module] && !requiredSet[module] {
			importedSet[module] = true

			if dependencySet[module] {
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

func parseModuleFromImportStmt(i string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("could not parse import %v with panic \n %v", i, r)
		}
	}()

	startIndex := strings.Index(i, "\"")
	endIndex := strings.LastIndex(i, "\"")
	if startIndex == -1 || endIndex == -1 || startIndex == endIndex {
		startIndex = strings.Index(i, "'")
		endIndex = strings.LastIndex(i, "'")
		if startIndex == -1 || endIndex == -1 {
			log.Printf("could not parse import %v", i)
			return ""
		}
	}

	module := i[startIndex+1 : endIndex]
	return module
}

func IsLocalImport(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

func stripQuotes(s string) string {
	return strings.Trim(s, `"'`)
}

func parsePackageJSON(packagePath string) (MinimalPackage, error) {
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
				log.Print(errors.Wrap(err, "error finding package json").Error())
				return packageMetadata, nil
			}
		}
	}

	fileContents, err := ioutil.ReadFile(pathToPackageJSON)
	if err != nil {
		log.Print(errors.Wrap(err, "error reading package json").Error())
		return packageMetadata, nil
	}

	// remove BOM prefix see https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	fileContents = bytes.TrimPrefix(fileContents, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(fileContents, &packageMetadata)
	if err != nil {
		log.Print(string(fileContents))
		log.Print(errors.Wrap(err, "error parsing package json").Error())
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
