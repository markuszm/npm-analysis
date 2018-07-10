package evolution

import (
	"io/ioutil"
	"net/http"
	"strings"
)

const npmUrl = "https://registry.npmjs.com/"

const replicateUrl = "https://replicate.npmjs.com/"

func GetMetadataFromNpm(pkg string) (string, error) {
	pkgName := pkg
	if strings.Contains(pkg, "/") {
		pkgName = transformScopedName(pkg)
	}
	url := npmUrl + pkgName

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	doc := string(bytes)
	return doc, err
}

func PackageStillExists(pkg string) (bool, error) {
	pkgName := pkg
	if strings.Contains(pkg, "/") {
		pkgName = transformScopedName(pkg)
	}
	url := replicateUrl + pkgName

	resp, err := http.Head(url)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
