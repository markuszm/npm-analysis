package cmd

import (
	"encoding/json"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var cgQueryNeo4jUrl string
var cgQueryMysqlUrl string
var cgQueryOutput string
var cgQueryInputFile string

// callgraphCmd represents the callgraph command
var cgQueryCmd = &cobra.Command{
	Use:   "cgQuery",
	Short: "Queries package callgraph",
	Long:  `Queries callgraph using neo4j database`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		queries, err := packagecallgraph.NewGraphQueries(cgQueryNeo4jUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer queries.Close()

		functionChan := make(chan string, 0)

		if cgQueryInputFile == "" {
			go queries.StreamExportedFunctions(functionChan)
		} else {
			go streamFunctionNamesFromFile(functionChan)
		}

		count := 0

		file, err := os.Create(cgQueryOutput)
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
	logger.Infow("using package input file", "file", cgQueryInputFile)
	file, err := ioutil.ReadFile(cgQueryInputFile)
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
	rootCmd.AddCommand(cgQueryCmd)

	cgQueryCmd.Flags().StringVarP(&cgQueryNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7689", "Neo4j bolt url")
	cgQueryCmd.Flags().StringVarP(&cgQueryMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	cgQueryCmd.Flags().StringVarP(&cgQueryOutput, "output", "o", "/home/markus/npm-analysis/apiUsage.json", "output file")
	cgQueryCmd.Flags().StringVarP(&cgQueryInputFile, "input", "i", "", "input file containing list with full function names")
}

type ApiUsageStats struct {
	FunctionName  string
	CallCount     int64
	PackagesCount int
	Packages      []string
}
