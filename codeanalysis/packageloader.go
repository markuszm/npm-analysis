package codeanalysis

import (
	"errors"
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"os"
	"path"
	"strings"
)

type PackageLoader interface {
	LoadPackage(pkg model.PackageVersionPair) (string, error)
	LoadPackages(packages []model.PackageVersionPair) (map[string]string, error)
}

const ErrorNotFound = util.ConstantError("package not found")

type WebLoader struct {
	RegistryUrl string
	TempFolder  string
}

func NewWebLoader(registryUrl string, tempFolder string) *WebLoader {
	return &WebLoader{RegistryUrl: registryUrl, TempFolder: tempFolder}
}

func (w *WebLoader) LoadPackage(packageName string) (string, error) {
	// TODO: if needed implement web loader
	return "", errors.New("not implemented")
}

func (w *WebLoader) LoadPackages(packageNames []model.PackageVersionPair) (map[string]string, error) {
	// TODO: if needed implement web loader
	result := make(map[string]string, len(packageNames))
	return result, errors.New("not implemented")
}

type DiskLoader struct {
	Path string
}

func NewDiskLoader(path string) (*DiskLoader, error) {
	return &DiskLoader{Path: path}, nil
}

func (d *DiskLoader) LoadPackage(pkg model.PackageVersionPair) (string, error) {
	p := path.Join(d.Path, GetPackageFilePath(pkg.Name, pkg.Version))
	if _, err := os.Stat(p); err != nil {
		// file does not exist
		return "", ErrorNotFound
	}
	return p, nil
}

func (d *DiskLoader) LoadPackages(packageNames []model.PackageVersionPair) (map[string]string, error) {
	result := make(map[string]string, len(packageNames))
	for _, pkg := range packageNames {
		pkgPath, err := d.LoadPackage(pkg)
		if err != nil {
			return result, err
		}
		result[pkg.Name] = pkgPath
	}
	return result, nil
}

func GetPackageFilePath(pkg, version string) string {
	fragments := strings.Split(pkg, "/")
	fileName := pkg
	if len(fragments) > 1 {
		scopedName := fragments[0]
		if strings.HasPrefix(scopedName, "@") {
			fileName = scopedName + "_" + fragments[1]
		}
	}
	fileName += fmt.Sprintf("-%v.tgz", version)

	firstLetter := string(fileName[0])
	secondLetter := ""
	if len(fileName) > 1 {
		secondLetter = string(fileName[1])
	}
	return path.Join(firstLetter, secondLetter, fileName)
}
