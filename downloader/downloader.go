package downloader

import (
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func DownloadPackage(downloadPath string, registryUrl string, pkg model.PackageVersionPair) (string, error) {
	downloadUrl := GenerateDownloadUrl(pkg, registryUrl)

	_, fileName, fileNameErr := GeneratePackageFileName(downloadUrl)
	if fileNameErr != nil {
		return "", errors.Wrapf(fileNameErr, "Error generating package filename: %s", downloadPath)
	}

	fullPath := path.Join(downloadPath, fileName)

	resp, requestErr := http.Get(downloadUrl)
	if requestErr != nil {
		return "", errors.Wrapf(requestErr, "Error downloading package: %s", downloadUrl)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("Not Found")
	}

	file, createFileErr := os.Create(fullPath)

	defer file.Close()

	if createFileErr != nil {
		return "", errors.Wrapf(createFileErr, "Error downloading package: %s", downloadUrl)
	}

	_, copyErr := io.Copy(file, resp.Body)

	if copyErr != nil {
		return "", errors.Wrapf(copyErr, "Error downloading package: %s", downloadUrl)
	}

	return fullPath, nil
}

func DownloadPackageAndVerify(downloadPath, pkgUrl, pkgShasum string) (string, error) {
	if !strings.Contains(pkgUrl, "registry.npmjs.org") {
		return "", errors.New("Not Found")
	}

	packageFileName, scopedFileName, fileNameErr := GeneratePackageFileName(pkgUrl)
	if fileNameErr != nil {
		return "", errors.Wrapf(fileNameErr, "Error generating package filename: %s", pkgUrl)
	}

	nestedDir, mkDirErr := CreateNestedFolders(scopedFileName, downloadPath)
	if mkDirErr != nil {
		return "", errors.Wrapf(mkDirErr, "Could not create nested folders for %s", pkgUrl)
	}

	packageFilePath := path.Join(nestedDir, packageFileName)
	scopedFilePath := path.Join(nestedDir, scopedFileName)
	if _, err := os.Stat(scopedFilePath); err == nil {
		// path exists already - but check integrity
		// rename to package file name (because we change the name to a scoped name)
		os.Rename(scopedFilePath, packageFilePath)
		integrityErr := VerifyIntegrity(pkgShasum, packageFilePath)
		if integrityErr != nil {
			return "", errors.Wrapf(integrityErr, "Error downloading package: %s", pkgUrl)
		}
		// rename it back
		os.Rename(packageFilePath, scopedFilePath)
		return scopedFilePath, nil
	}
	resp, requestErr := http.Get(pkgUrl)
	if requestErr != nil {
		return "", errors.Wrapf(requestErr, "Error downloading package: %s", pkgUrl)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("Not Found")
	}

	file, createFileErr := os.Create(packageFilePath)

	defer file.Close()

	if createFileErr != nil {
		return "", errors.Wrapf(createFileErr, "Error downloading package: %s", pkgUrl)
	}

	_, copyErr := io.Copy(file, resp.Body)

	if copyErr != nil {
		return "", errors.Wrapf(copyErr, "Error downloading package: %s", pkgUrl)
	}

	integrityErr := VerifyIntegrity(pkgShasum, packageFilePath)
	if integrityErr != nil {
		return "", errors.Wrapf(integrityErr, "Error downloading package: %s", pkgUrl)
	}

	// rename to scoped name
	renameErr := os.Rename(packageFilePath, scopedFilePath)
	if renameErr != nil {
		log.Fatalf("cant rename package file to scoped name for package %s", pkgUrl)
	}
	return scopedFilePath, nil
}

func CreateNestedFolders(fileName, downloadPath string) (string, error) {
	firstLetter := string(fileName[0])
	secondLetter := ""
	if len(fileName) > 1 {
		secondLetter = string(fileName[1])
	}

	nestedDir := path.Join(downloadPath, firstLetter, secondLetter)
	mkDirErr := os.MkdirAll(nestedDir, os.ModePerm)
	if mkDirErr != nil {
		return "", mkDirErr
	}
	return nestedDir, nil
}
