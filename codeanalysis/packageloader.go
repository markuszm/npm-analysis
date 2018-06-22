package codeanalysis

import (
	"fmt"
	"github.com/markuszm/npm-analysis/downloader"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"os"
	"path"
	"strings"
)

type PackageLoader interface {
	LoadPackage(pkg model.PackageVersionPair) (string, error)
	LoadPackages(packages []model.PackageVersionPair) (map[string]string, error)
	NeedsCleanup() bool
}

const ErrorNotFound = util.ConstantError("package not found")

type NetLoader struct {
	RegistryUrl string
	TempFolder  string
}

func NewNetLoader(registryUrl string, tempFolder string) *NetLoader {
	return &NetLoader{RegistryUrl: registryUrl, TempFolder: tempFolder}
}

func (n *NetLoader) LoadPackage(pkg model.PackageVersionPair) (string, error) {
	dlPath, err := downloader.DownloadPackage(n.TempFolder, n.RegistryUrl, pkg)
	if err != nil {
		if err.Error() == "Not Found" {
			return "", ErrorNotFound
		}
		return "", err
	}
	return dlPath, nil
}

func (n *NetLoader) LoadPackages(packages []model.PackageVersionPair) (map[string]string, error) {
	result := make(map[string]string, len(packages))
	for _, pkg := range packages {
		pkgPath, err := n.LoadPackage(pkg)
		if err != nil {
			return result, err
		}
		result[pkg.Name] = pkgPath
	}
	return result, nil
}

func (n *NetLoader) NeedsCleanup() bool {
	return true
}

type DiskLoader struct {
	Path string
}

func NewDiskLoader(path string) *DiskLoader {
	return &DiskLoader{Path: path}
}

func (d *DiskLoader) LoadPackage(pkg model.PackageVersionPair) (string, error) {
	p := path.Join(d.Path, GetPackageFilePath(pkg.Name, pkg.Version))
	if _, err := os.Stat(p); err != nil {
		// file does not exist
		return "", ErrorNotFound
	}
	return p, nil
}

func (d *DiskLoader) LoadPackages(packages []model.PackageVersionPair) (map[string]string, error) {
	result := make(map[string]string, len(packages))
	for _, pkg := range packages {
		pkgPath, err := d.LoadPackage(pkg)
		if err != nil {
			return result, err
		}
		result[pkg.Name] = pkgPath
	}
	return result, nil
}

func (d *DiskLoader) NeedsCleanup() bool {
	return false
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
