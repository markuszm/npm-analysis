package downloader

import (
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"net/url"
	"path"
	"strings"
)

func GeneratePackageFileName(downloadUrl string) (string, string, error) {
	parsedUrl, parseErr := url.Parse(downloadUrl)
	if parseErr != nil {
		return "", "", parseErr
	}
	pathFragments := strings.Split(parsedUrl.Path, "/")
	// take scoped name concatenated with file name
	scopedName := pathFragments[1]
	packageFileName := path.Base(parsedUrl.Path)
	scopedFileName := packageFileName
	if strings.HasPrefix(scopedName, "@") {
		scopedFileName = scopedName + "_" + packageFileName
	}
	return packageFileName, scopedFileName, nil
}

func GeneratePackageFullPath(downloadUrl string) (string, error) {
	if downloadUrl == "" {
		return "", errors.New("empty download url")
	}

	_, fileName, err := GeneratePackageFileName(downloadUrl)
	if err != nil {
		return "", err
	}

	firstLetter := string(fileName[0])
	secondLetter := ""
	if len(fileName) > 1 {
		secondLetter = string(fileName[1])
	}

	fullPath := path.Join(firstLetter, secondLetter, fileName)
	return fullPath, nil
}

func GenerateDownloadUrl(pkg model.PackageVersionPair, registryUrl string) string {
	fragments := strings.Split(pkg.Name, "/")
	packageName := pkg.Name
	if len(fragments) > 1 {
		scopedName := fragments[0]
		if strings.HasPrefix(scopedName, "@") {
			packageName = fragments[1]
		}
	}
	pkgUrl := fmt.Sprintf("%v/%v/-/%v-%v.tgz", registryUrl, pkg.Name, packageName, pkg.Version)
	return pkgUrl
}
