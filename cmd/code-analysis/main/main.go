package main

import (
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysispipeline"
	"github.com/markuszm/npm-analysis/codeanalysispipeline/codeanalysis"
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
	resultPath := flag.String("result", "/home/markus/npm-analysis/code-analysis.json", "File path to store results in")

	collectorFlag := flag.String("collector", "db", "how to collect package names (db or file)")
	file := flag.String("namesFile", "./codeanalysispipeline/testfiles/test-packages.txt", "filepath containing package names")

	loaderFlag := flag.String("loader", "disk", "specify loader type (disk or net)")
	registryUrl := flag.String("registry", "http://registry.npmjs.org", "npm registry url (only when using net loader)")

	writerFlag := flag.String("writer", "json", "specify writer type (csv or json)")

	analysisFlag := flag.String("analysis", "file_distribution", "specify which analysis to run")

	flag.Parse()

	var collector codeanalysispipeline.NameCollector
	switch *collectorFlag {
	case "db":
		log.Print("using db collector")
		var err error
		collector, err = codeanalysispipeline.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", mysqlUser, mysqlPw))
		if err != nil {
			log.Fatal(err)
		}
	case "file":
		log.Printf("using file collector with file %v", *file)
		collector = codeanalysispipeline.NewFileNameCollector(*file)
	}

	//collector := codeanalysispipeline.NewTestNameCollector([]model.PackageVersionPair{
	//	{Name: "1720", Version: "1.0.0"},
	//	{Name: "@esm/ms", Version: "2.0.1"},
	//	{Name: "@fav/path", Version: "0.9.0"},
	//})

	var loader codeanalysispipeline.PackageLoader

	switch *loaderFlag {
	case "disk":
		log.Printf("using disk loader from packages path %v", *packagesPath)
		loader = codeanalysispipeline.NewDiskLoader(*packagesPath)
	case "net":
		log.Printf("using net loader from registry %v and storing temp packages into path %v", *registryUrl, *tmpPath)
		loader = codeanalysispipeline.NewNetLoader(*registryUrl, *tmpPath)
	}

	unpacker := codeanalysispipeline.NewDiskUnpacker(*tmpPath)

	var analysis codeanalysis.AnalysisExecutor
	switch *analysisFlag {
	case "file_distribution":
		log.Print("executing file distribution analysis")
		analysis = &codeanalysis.FileDistributionAnalysis{}
	case "used_dependencies":
		log.Print("executing used dependencies analysis")
		analysis = &codeanalysis.UsedDependenciesAnalysis{}
	}

	var writer codeanalysispipeline.ResultWriter
	switch *writerFlag {
	case "csv":
		log.Printf("using csv result writer to path %v", *resultPath)
		writer = codeanalysispipeline.NewCSVWriter(*resultPath)
	case "json":
		log.Printf("using json result writer to path %v", *resultPath)
		writer = codeanalysispipeline.NewJSONWriter(*resultPath)
	}

	pipeline := codeanalysispipeline.NewPipeline(collector, loader, unpacker, analysis, writer)

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
