package cmd

import (
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
)

var callgraphInputCallgraph string
var callgraphInputExports string
var callgraphNeo4jUrl string
var callgraphWorkerNumber int

// callgraphCmd represents the callgraph command
var callgraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "Generates callgraph",
	Long:  `Generates callgraph using ast analysis results as input and writes it to neo4j database`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		graphCreator := packagecallgraph.NewGraphCreator(callgraphNeo4jUrl, callgraphInputCallgraph, callgraphInputExports, callgraphWorkerNumber, logger)
		err := graphCreator.ExecCreation()
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(callgraphCmd)

	callgraphCmd.Flags().StringVarP(&callgraphInputCallgraph, "callgraph", "c", "", "Path to callgraph analysis results")
	callgraphCmd.Flags().StringVarP(&callgraphInputExports, "exports", "e", "", "Path to exports analysis results")
	callgraphCmd.Flags().StringVarP(&callgraphNeo4jUrl, "url", "u", "bolt://neo4j:npm@localhost:7688", "Neo4j bolt url")
	callgraphCmd.Flags().IntVarP(&callgraphWorkerNumber, "worker", "w", 10, "Number of workers")
}
