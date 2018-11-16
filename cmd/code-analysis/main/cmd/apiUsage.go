package cmd

import (
	"encoding/json"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/markuszm/npm-analysis/codeanalysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var apiUsageNeo4jUrl string
var apiUsageMysqlUrl string
var apiUsageOutput string
var apiUsageInputFile string

// callgraphCmd represents the callgraph command
var apiUsageCmd = &cobra.Command{
	Use:   "apiUsage",
	Short: "Generates api usage stats for functions",
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		queries, err := packagecallgraph.NewGraphQueries(apiUsageNeo4jUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer queries.Close()

		functionChan := make(chan string, 0)

		if apiUsageInputFile == "" {
			go queries.StreamExportedFunctions("actualExport", functionChan)
		} else {
			go streamFunctionNamesFromFile(functionChan)
		}

		count := 0

		file, err := os.Create(apiUsageOutput)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)

		for expFunc := range functionChan {
			callCount, err := queries.GetCallCountForExportedFunction(expFunc)
			if err != nil {
				log.Fatal(err)
			}

			packages, err := queries.GetPackagesThatCallExportedFunction(expFunc)
			if err != nil {
				log.Fatal(err)
			}

			apiUsageStat := ApiUsageStats{
				FunctionName:  expFunc,
				CallCount:     callCount,
				Packages:      packages,
				PackagesCount: len(packages),
			}

			err = encoder.Encode(apiUsageStat)
			if err != nil {
				log.Fatal("Cannot write to result file", err)
			}

			count++

			if count%10000 == 0 {
				logger.Infof("Finished %v functions", count)
			}
		}

	},
}

func streamFunctionNamesFromFile(functionChan chan string) {
	logger.Infow("using package input file", "file", apiUsageInputFile)
	file, err := ioutil.ReadFile(apiUsageInputFile)
	if err != nil {
		logger.Fatalw("could not read file", "err", err)
	}
	lines := strings.Split(string(file), "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		functionChan <- l
	}

	close(functionChan)
}

func init() {
	rootCmd.AddCommand(apiUsageCmd)

	apiUsageCmd.Flags().StringVarP(&apiUsageNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7689", "Neo4j bolt url")
	apiUsageCmd.Flags().StringVarP(&apiUsageMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	apiUsageCmd.Flags().StringVarP(&apiUsageOutput, "output", "o", "/home/markus/npm-analysis/apiUsage.json", "output file")
	apiUsageCmd.Flags().StringVarP(&apiUsageInputFile, "input", "i", "", "input file containing list with full function names")
}

type ApiUsageStats struct {
	FunctionName  string
	CallCount     int64
	PackagesCount int
	Packages      []string
}
