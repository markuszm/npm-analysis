package resultprocessing

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
)

type Export struct {
	ExportType string   `json:"type"`
	Identifier string   `json:"id"`
	Arguments  []string `json:"args"`
	BundleType string   `json:"bundleType"`
	File       string   `json:"file"`
	IsDefault  bool     `json:"isDefault"`
	Local      string   `json:"local"`
}

func TransformToExports(result interface{}) ([]Export, error) {
	if reflect.TypeOf(result).String() != "[]interface {}" {
		return nil, errors.New("Error parsing exports")
	}

	objs := result.([]interface{})

	var exports []Export

	for _, value := range objs {
		export := Export{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return exports, err
		}
		err = json.Unmarshal(bytes, &export)
		if err != nil {
			return exports, err
		}
		exports = append(exports, export)
	}
	return exports, nil
}
