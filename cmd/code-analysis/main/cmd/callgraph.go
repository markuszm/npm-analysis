package cmd

import (
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
)

var callgraphInputCallgraph string
var callgraphInputExports string
var callgraphNeo4jUrl string
var callgraphWorkerNumber int

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

// callgraphCmd represents the callgraph command
var callgraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "Generates callgraph",
	Long:  `Generates callgraph using ast analysis results as input and writes it to neo4j database`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		var mysqlUrl = fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW)

		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(mysqlUrl)
		defer mysql.Close()
		if databaseInitErr != nil {
			logger.Fatal(databaseInitErr)
		}

		err := packagecallgraph.InitSchema(callgraphNeo4jUrl)
		if err != nil {
			logger.Fatal(err)
		}

		callEdgeCreator := packagecallgraph.NewCallEdgeCreator(callgraphNeo4jUrl, callgraphInputCallgraph, callgraphWorkerNumber, mysql, logger)
		err = callEdgeCreator.Exec()
		if err != nil {
			logger.Fatal(err)
		}

		exportEdgeCreator := packagecallgraph.NewExportEdgeCreator(callgraphNeo4jUrl, callgraphInputExports, callgraphWorkerNumber, logger)
		err = exportEdgeCreator.Exec()
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
