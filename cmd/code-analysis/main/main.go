package main

import (
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/codeanalysis/analysisimpl"
	"io/ioutil"
	"log"
	"os"
)

const mysqlUser = "root"
const mysqlPw = "npm-analysis"

const packagesPath = "/media/markus/NPM/NPM"
const tmpPath = "/home/markus/tmp"
const resultPath = "/home/markus/npm-analysis/code-analysis.json"
const maxWorkers = 100

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

	formatter := &codeanalysis.JSONFormatter{}

	pipeline := codeanalysis.NewPipeline(collector, loader, unpacker, analysis, formatter)

	var result string
	if *parallel {
		log.Printf("Running in parallel with %v workers", maxWorkers)
		result, err = pipeline.ExecuteParallel(maxWorkers)
	} else {
		log.Print("Running in sequential mode")
		result, err = pipeline.Execute()
	}

	if err != nil {
		log.Printf("ERRORS: \n %v", err)
	}

	ioutil.WriteFile(resultPath, []byte(result), os.ModePerm)

	log.Printf("Wrote results to %v", resultPath)

	// TODO on exit cleanup
}
