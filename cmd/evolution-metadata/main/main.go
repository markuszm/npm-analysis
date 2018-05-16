package main

import (
	"github.com/markuszm/npm-analysis/database/evolution"
	"log"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

func main() {
	mongoDB := evolution.NewMongoDB(MONGOURL, "npm", "packages")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	doc, err := mongoDB.FindOneSimple("key", "0")
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	val, err := evolution.Decompress(doc)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	log.Printf("Value: %v", val)
}
