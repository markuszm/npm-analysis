package main

import (
	"io/ioutil"
	"github.com/buger/jsonparser"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm-all.json"

func main() {
	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//jsonparser.ObjectEach(value, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		//	// Here you can access each object value, if ValueType is object you need to nest further
		//	return nil
		//}, "value")
		//
		id, _ := jsonparser.GetString(value, "id")
		dep, _,_, _ := jsonparser.Get(value, "value", "dependencies")
		println(id, string(dep))
	}, "rows")

}
