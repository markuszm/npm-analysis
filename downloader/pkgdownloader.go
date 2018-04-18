package downloader

import (
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func DownloadPackage(downloadPath, url string) error {
	fileName, fileNameErr := GeneratePackageFileName(url)
	if fileNameErr != nil {
		return errors.Wrapf(fileNameErr, "Error generating package filename: %s", url)
	}

	filePath := path.Join(downloadPath, fileName)
	if _, err := os.Stat(filePath); err == nil {
		// path exists already - skip
		return nil
	}
	resp, requestErr := http.Get(url)
	if requestErr != nil {
		return errors.Wrapf(requestErr, "Error downloading package: %s", url)
	}

	defer resp.Body.Close()

	file, createFileErr := os.Create(filePath)

	defer file.Close()

	if createFileErr != nil {
		return errors.Wrapf(createFileErr, "Error downloading package: %s", url)
	}

	_, copyErr := io.Copy(file, resp.Body)

	if copyErr != nil {
		return errors.Wrapf(copyErr, "Error downloading package: %s", url)
	}

	return nil
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
