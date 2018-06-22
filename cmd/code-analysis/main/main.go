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

//const packagesPath = "/home/markus/npm-analysis/test"

func main() {
	parallel := flag.Bool("parallel", false, "Execute pipeline in parallel?")
	maxWorkers := flag.Int("workers", 100, "number of workers (only if parallel)")

	packagesPath := flag.String("packages", "/media/markus/NPM/NPM", "folder path to packages")
	tmpPath := flag.String("tmp", "/home/markus/tmp", "Temp path to store extracted packages")
	resultPath := flag.String("result", "/home/markus/npm-analysis/code-analysis.csv", "File path to store results in")

	collectorFlag := flag.String("collector", "db", "how to collect package names (db or file)")
	file := flag.String("namesFile", "./codeanalysis/testfiles/test-packages.txt", "filepath containing package names")

	loaderFlag := flag.String("loader", "disk", "specify loader type (disk or net)")
	registryUrl := flag.String("registry", "http://registry.npmjs.org", "npm registry url (only when using net loader)")

	writerFlag := flag.String("writer", "csv", "specify writer type (csv or json)")

	flag.Parse()

	var collector codeanalysis.NameCollector
	switch *collectorFlag {
	case "db":
		var err error
		collector, err = codeanalysis.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", mysqlUser, mysqlPw))
		if err != nil {
			log.Fatal(err)
		}
	case "file":
		collector = codeanalysis.NewFileNameCollector(*file)
	}

	//collector := codeanalysis.NewTestNameCollector([]model.PackageVersionPair{
	//	{Name: "1720", Version: "1.0.0"},
	//	{Name: "@esm/ms", Version: "2.0.1"},
	//	{Name: "@fav/path", Version: "0.9.0"},
	//})

	var loader codeanalysis.PackageLoader

	switch *loaderFlag {
	case "disk":
		loader = codeanalysis.NewDiskLoader(*packagesPath)
	case "net":
		loader = codeanalysis.NewNetLoader(*registryUrl, *tmpPath)
	}

	unpacker := codeanalysis.NewDiskUnpacker(*tmpPath)

	analysis := &analysisimpl.FileDistributionAnalysis{}

	var writer codeanalysis.ResultWriter
	switch *writerFlag {
	case "csv":
		writer = codeanalysis.NewCSVWriter(*resultPath)
	case "json":
		writer = codeanalysis.NewJSONWriter(*resultPath)
	}

	pipeline := codeanalysis.NewPipeline(collector, loader, unpacker, analysis, writer)

	var err error
	if *parallel {
		log.Printf("Running in parallel with %v workers", *maxWorkers)
		err = pipeline.ExecuteParallel(*maxWorkers)
	} else {
		log.Print("Running in sequential mode")
		err = pipeline.Execute()
	}

	if err != nil {
		log.Printf("ERRORS: \n %v", err)
	}

	// TODO on exit cleanup
}
