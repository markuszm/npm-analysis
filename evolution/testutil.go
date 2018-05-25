package evolution

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"io/ioutil"
	"testing"
)

const lodashTestJsonPath = "./testfiles/lodash-all.json"
const expressTestJsonPath = "./testfiles/express-all.json"
const devCliTestJsonPath = "./testfiles/@anycli-dev-cli.json"

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
