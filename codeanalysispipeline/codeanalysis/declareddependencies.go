package codeanalysis

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type UsedDependenciesAnalysis struct {
}

func (d *UsedDependenciesAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	requiredResult, err := ExecuteCommand("grep", `--include=*.js`, `--exclude-dir=node_modules`, "-ohrP", `(?<=require\().+?(?=\))`, packagePath)
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 1") {
			return nil, errors.Wrap(err, "error retrieving required packages")
		}
	}

	packageMetadata, err := parsePackageJSON(packagePath)

	if err != nil {
		return nil, errors.Wrap(err, "error parsing package json")
	}

	requiredPackages := strings.Split(requiredResult, "\n")

	requiredSet := make(map[string]bool, 0)
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
		if module != "" && !build.IsLocalImport(module) && !requiredSet[module] {
			requiredSet[module] = true

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

	dependencies := make([]string, len(dependencySet))
	i = 0
	for k := range dependencySet {
		dependencies[i] = k
		i++
	}

	result := DependencyResult{
		Required:     required,
		Dependencies: dependencies,
		Used:         used,
	}

	return result, nil
}

func stripQuotes(s string) string {
	return strings.Trim(s, `"'`)
}

func parsePackageJSON(packagePath string) (model.Package, error) {
	var packageMetadata model.Package
	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "error finding package json")
		}
		if info.Name() == "package.json" {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, "error reading package json")
			}

			err = json.Unmarshal(bytes, &packageMetadata)
			if err != nil {
				return errors.Wrap(err, "error parsing package json")
			}
		}
		return nil
	})
	return packageMetadata, err
}

type DependencyResult struct {
	Required     []string
	Dependencies []string
	Used         []string
}
