package resultprocessing

import (
	"encoding/json"
	"errors"
	"reflect"
)

type DynamicExport struct {
	Name         string     `json:"Name"`
	InternalName string     `json:"InternalName"`
	Locations    []Location `json:"Locations"`
}

type Location struct {
	File  string `json:"File"`
	Index int    `json:"Index"`
}

func TransformToDynamicExports(result interface{}) ([]DynamicExport, error) {
	if reflect.TypeOf(result).String() != "[]interface {}" {
		return nil, errors.New("error parsing dynamic exports")
	}

	objs := result.([]interface{})

	var dynamicExports []DynamicExport

	for _, value := range objs {
		export := DynamicExport{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return dynamicExports, err
		}
		err = json.Unmarshal(bytes, &export)
		if err != nil {
			return dynamicExports, err
		}
		dynamicExports = append(dynamicExports, export)
	}
	return dynamicExports, nil
}
