package main

import (
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/codeanalysis/analysisimpl"
	"log"
)

const mysqlUser = "root"
const mysqlPw = "npm-analysis"

const packagesPath = "/media/markus/NPM/NPM"

//const packagesPath = "/home/markus/npm-analysis/test"
const tmpPath = "/home/markus/tmp"
const resultPath = "/home/markus/npm-analysis/code-analysis.csv"
const maxWorkers = 1000

func main() {
	parallel := flag.Bool("parallel", false, "Execute pipeline in parallel?")
	flag.Parse()

	collector, err := codeanalysis.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", mysqlUser, mysqlPw))
	if err != nil {
		log.Fatal(err)
	}

	//collector, err := codeanalysis.NewFileNameCollector("./codeanalysis/testfiles/test-packages.txt")
	//if err != nil {
	//	log.Fatal(err)
	//}

	//collector := codeanalysis.NewTestNameCollector([]model.PackageVersionPair{
	//	{Name: "1720", Version: "1.0.0"},
	//	{Name: "@esm/ms", Version: "2.0.1"},
	//	{Name: "@fav/path", Version: "0.9.0"},
	//})

	loader, err := codeanalysis.NewDiskLoader(packagesPath)
	if err != nil {
		log.Fatal(err)
	}

	unpacker := codeanalysis.NewDiskUnpacker(tmpPath)

	analysis := &analysisimpl.EmptyPackageAnalysis{}

	writer := codeanalysis.NewCSVWriter(resultPath)

	pipeline := codeanalysis.NewPipeline(collector, loader, unpacker, analysis, writer)

	if *parallel {
		log.Printf("Running in parallel with %v workers", maxWorkers)
		err = pipeline.ExecuteParallel(maxWorkers)
	} else {
		log.Print("Running in sequential mode")
		err = pipeline.Execute()
	}

	if err != nil {
		log.Printf("ERRORS: \n %v", err)
	}

	// TODO on exit cleanup
}
