package evolution

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"io/ioutil"
	"testing"
	"time"
)

const lodashTestJsonPath = "./testfiles/lodash-all.json"
const expressTestJsonPath = "./testfiles/express-all.json"
const reactTestJsonPath = "./testfiles/react.json"
const devCliTestJsonPath = "./testfiles/@anycli-dev-cli.json"

const lodashDownloadCountsPath = "./testfiles/lodash-downloadcounts.json"

var timeCutoff = time.Unix(1523626680, 0)

func MustReadMetadataFromTestFile(testFilePath string, t *testing.T) model.Metadata {
	bytes, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	var testPackage model.Metadata
	err = json.Unmarshal(bytes, &testPackage)
	if err != nil {
		t.Fatal(err)
	}
	return testPackage
}

func MustReadDownloadCountsFromTestFile(testFilePath string, t *testing.T) DownloadCountResponse {
	bytes, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	var downloadCounts DownloadCountResponse
	err = json.Unmarshal(bytes, &downloadCounts)
	if err != nil {
		t.Fatal(err)
	}
	return downloadCounts
}
