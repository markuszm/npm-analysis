package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"npm-analysis/database/model"
	"os"
	"path"
	"strings"
)

func DownloadPackage(downloadPath string, pkg model.Dist) error {
	pkgUrl := pkg.Url
	packageFileName, scopedFileName, fileNameErr := GeneratePackageFileName(pkgUrl)
	if fileNameErr != nil {
		return errors.Wrapf(fileNameErr, "Error generating package filename: %s", pkgUrl)
	}

	packageFilePath := path.Join(downloadPath, packageFileName)
	scopedFilePath := path.Join(downloadPath, scopedFileName)
	if _, err := os.Stat(scopedFilePath); err == nil {
		// path exists already - but check integrity
		// rename to package file name (because we change the name to a scoped name)
		os.Rename(scopedFilePath, packageFilePath)
		integrityErr := VerifyIntegrity(pkg.Shasum, packageFilePath)
		if integrityErr != nil {
			return errors.Wrapf(integrityErr, "Error downloading package: %s", pkgUrl)
		}
		// rename it back
		os.Rename(packageFilePath, scopedFilePath)
		return nil
	}
	resp, requestErr := http.Get(pkgUrl)
	if requestErr != nil {
		return errors.Wrapf(requestErr, "Error downloading package: %s", pkgUrl)
	}

	defer resp.Body.Close()

	file, createFileErr := os.Create(packageFilePath)

	defer file.Close()

	if createFileErr != nil {
		return errors.Wrapf(createFileErr, "Error downloading package: %s", pkgUrl)
	}

	_, copyErr := io.Copy(file, resp.Body)

	if copyErr != nil {
		return errors.Wrapf(copyErr, "Error downloading package: %s", pkgUrl)
	}

	integrityErr := VerifyIntegrity(pkg.Shasum, packageFilePath)
	if integrityErr != nil {
		return errors.Wrapf(integrityErr, "Error downloading package: %s", pkgUrl)
	}

	// rename to scoped name
	renameErr := os.Rename(packageFilePath, scopedFilePath)
	if renameErr != nil {
		log.Fatalf("cant rename package file to scoped name for package %s", pkgUrl)
	}
	return nil
}

func VerifyIntegrity(shasum, filePath string) error {
	hasher := sha1.New()

	file, openErr := os.Open(filePath)
	defer file.Close()

	if openErr != nil {
		return errors.Wrapf(openErr, "Error opening package")
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return errors.Wrapf(err, "Error opening calculating shasum for package: %s", file.Name())
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if checksum != shasum {
		// delete file if it fails integrity check
		deleteErr := os.Remove(filePath)
		if deleteErr != nil {
			log.Fatal("Error deleting file with failed integrity check - must stop")
			return deleteErr
		}
		return errors.New("File integrity check failed")
	}

	return nil
}

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
