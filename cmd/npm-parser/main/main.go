package main

import (
	//"io/ioutil"
	//"github.com/buger/jsonparser"
	"npm-analysis/database"
	"log"
	"github.com/buger/jsonparser"
	"io/ioutil"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm-all.json"

const DATABASE_PATH = "/home/markus/npm-analysis/npm.db"

func main() {
	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		jsonparser.ObjectEach(value, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			// Here you can access each object value, if ValueType is object you need to nest further
			return nil
		}, "value")

		id, _ := jsonparser.GetString(value, "id")
		val, _,_, _ := jsonparser.Get(value, "value", "maintainers")
		valAsStr := string(val)
		if valAsStr != "" {
			println(id, valAsStr)
		}
	}, "rows")

	db, databaseInitErr := database.InitDB(DATABASE_PATH)

	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	tableCreationErr := database.CreateTables(db)
	if tableCreationErr != nil {
		log.Fatal(tableCreationErr)
	}

}
