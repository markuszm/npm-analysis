package cmd

import (
	"encoding/json"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var cgQueryNeo4jUrl string
var cgQueryMysqlUrl string
var cgQueryOutput string

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

		go queries.StreamExportedFunctions(functionChan)

		var apiUsages []ApiUsageStats

		count := 0

		for expFunc := range functionChan {
			callCount, err := queries.GetCallCountForExportedFunction(expFunc)
			if err != nil {
				log.Fatal(err)
			}

			packages, err := queries.GetPackagesThatCallExportedFunction(expFunc)
			if err != nil {
				log.Fatal(err)
			}

			apiUsages = append(apiUsages, ApiUsageStats{
				FunctionName:  expFunc,
				CallCount:     callCount,
				Packages:      packages,
				PackagesCount: len(packages),
			})

			count++

			if count%10000 == 0 {
				logger.Infof("Finished %v functions", count)
			}
		}

		marshal, err := json.Marshal(apiUsages)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(cgQueryOutput, marshal, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cgQueryCmd)

	cgQueryCmd.Flags().StringVarP(&cgQueryNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7688", "Neo4j bolt url")
	cgQueryCmd.Flags().StringVarP(&cgQueryMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	cgQueryCmd.Flags().StringVarP(&cgQueryOutput, "output", "o", "/home/markus/npm-analysis/apiUsage.json", "output file")
}

type ApiUsageStats struct {
	FunctionName  string
	CallCount     int64
	PackagesCount int
	Packages      []string
}
