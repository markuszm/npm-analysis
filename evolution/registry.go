package evolution

import (
	"io/ioutil"
	"net/http"
	"strings"
)

const npmUrl = "https://registry.npmjs.com/"

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

func transformScopedName(pkg string) string {
	return strings.Replace(pkg, "/", "%2f", -1)
}
