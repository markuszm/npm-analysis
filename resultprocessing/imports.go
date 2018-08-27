package resultprocessing

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
)

type Import struct {
	Identifier string `json:"id"`
	FromModule string `json:"fromModule"`
	ModuleName string `json:"moduleName"`
	BundleType string `json:"bundleType"`
	Imported   string `json:"imported"`
}

func TransformToImports(result interface{}) ([]Import, error) {
	if reflect.TypeOf(result).String() != "[]interface {}" {
		return nil, errors.New("Error parsing imports")
	}

	objs := result.([]interface{})

	var imports []Import

	for _, value := range objs {
		importObj := Import{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return imports, err
		}
		err = json.Unmarshal(bytes, &importObj)
		if err != nil {
			return imports, err
		}
		imports = append(imports, importObj)
	}
	return imports, nil
}
