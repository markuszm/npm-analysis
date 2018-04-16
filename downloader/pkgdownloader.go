package downloader

import (
	"log"
	"github.com/pkg/errors"
	"path"
	"os"
	"net/http"
	"io"
	"strings"
	"net/url"
)

func DownloadPackage(downloadPath, url string) {
	fileName, fileNameErr := GeneratePackageFileName(url)
	if fileNameErr != nil {
		log.Fatal(errors.Wrapf(fileNameErr, "Error downloading package: %s", url))
	}

	filePath := path.Join(downloadPath, fileName)
	if _, err := os.Stat(filePath); err == nil {
		// path exists already - skip
		return
	}
	resp, requestErr := http.Get(url)
	if requestErr != nil {
		log.Fatal(errors.Wrapf(requestErr, "Error downloading package: %s", url))
	}

	defer resp.Body.Close()

	file, createFileErr := os.Create(filePath)

	defer file.Close()

	if createFileErr != nil {
		log.Fatal(errors.Wrapf(createFileErr, "Error downloading package: %s", url))
	}

	_, copyErr := io.Copy(file, resp.Body)

	if copyErr != nil {
		log.Fatal(errors.Wrapf(copyErr, "Error downloading package: %s", url))
	}

}

func GeneratePackageFileName(downloadUrl string) (string, error) {
	parsedUrl, parseErr := url.Parse(downloadUrl)
	if parseErr != nil {
		return "", parseErr
	}
	pathFragments := strings.Split(parsedUrl.Path, "/")
	// take scoped name concatenated with file name
	scopedName := pathFragments[1]
	packageFileName := path.Base(parsedUrl.Path)
	fileName := packageFileName
	if strings.HasPrefix(scopedName, "@") {
		fileName = scopedName + "_" + packageFileName
	}
	return fileName, nil
}

