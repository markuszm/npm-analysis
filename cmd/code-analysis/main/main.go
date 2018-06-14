package main

import (
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"log"
)

const MYSQL_USER = "root"
const MYSQL_PW = "npm-analysis"

func main() {
	collector, err := codeanalysis.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if err != nil {
		log.Fatal(err)
	}

	pipeline := codeanalysis.NewPipeline(collector)
	pipeline.Execute()
}
