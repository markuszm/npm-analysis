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

const packagesPath = ""
const tmpPath = ""
const resultPath = ""

func main() {
	parallel := flag.Bool("parallel", false, "Execute pipeline in parallel?")
	flag.Parse()

	collector, err := codeanalysis.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", mysqlUser, mysqlPw))
	if err != nil {
		log.Fatal(err)
	}

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
		result, err = pipeline.ExecuteParallel(10)
	} else {
		result, err = pipeline.Execute()
	}

	if err != nil {
		log.Printf("ERRORS: \n %v", err)
	}

	ioutil.WriteFile(resultPath, []byte(result), os.ModePerm)
}
