package cmd

import (
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"os"
)

var callgraphInputCallgraph string
var callgraphInputExports string
var callgraphNeo4jUrl string
var callgraphWorkerNumber int
var callgraphMysqlUrl string
var callgraphCSVFolder string
var callgraphInitSchema bool
var callgraphDynamicExports bool

// callgraphCmd represents the callgraph command
var callgraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "Generates callgraph",
	Long:  `Generates callgraph using ast analysis results as input and writes it to neo4j database`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(callgraphMysqlUrl)
		if databaseInitErr != nil {
			logger.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		if callgraphInputCallgraph == "" {
			logger.Info("Skipping callgraph creation")
		} else {
			var callEdgeCreator packagecallgraph.EdgeCreator

			if callgraphCSVFolder != "" {
				err := os.MkdirAll(callgraphCSVFolder, os.ModePerm)
				if err != nil {
					logger.Fatalw("could not create output folder", "err", err)
				}

				callEdgeCreator = packagecallgraph.NewCallEdgeCreatorCSV(callgraphCSVFolder, callgraphInputCallgraph, callgraphWorkerNumber, mysql, logger)
			} else {
				err := packagecallgraph.InitSchema(callgraphNeo4jUrl)
				if err != nil {
					logger.Fatal(err)
				}

				callEdgeCreator = packagecallgraph.NewCallEdgeCreator(callgraphNeo4jUrl, callgraphInputCallgraph, callgraphWorkerNumber, mysql, logger)
			}

			err := callEdgeCreator.Exec()
			if err != nil {
				logger.Fatal(err)
			}
		}

		if callgraphInputExports == "" {
			logger.Info("Skipping export creation")
		} else {
			if callgraphInitSchema {
				err := packagecallgraph.InitSchema(callgraphNeo4jUrl)
				if err != nil {
					logger.Fatal(err)
				}
			}

			var exportEdgeCreator packagecallgraph.EdgeCreator
			if callgraphDynamicExports {
				exportEdgeCreator = packagecallgraph.NewDynamicExportEdgeCreator(callgraphNeo4jUrl, callgraphInputExports, 1, mysql, logger)
			} else {
				exportEdgeCreator = packagecallgraph.NewExportEdgeCreator(callgraphNeo4jUrl, callgraphInputExports, 1, logger)
			}
			err := exportEdgeCreator.Exec()
			if err != nil {
				logger.Fatal(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(callgraphCmd)

	callgraphCmd.Flags().StringVarP(&callgraphInputCallgraph, "callgraph", "c", "", "Path to callgraph analysis results")
	callgraphCmd.Flags().StringVarP(&callgraphInputExports, "exports", "e", "", "Path to exports analysis results")
	callgraphCmd.Flags().StringVarP(&callgraphNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7688", "Neo4j bolt url")
	callgraphCmd.Flags().StringVarP(&callgraphMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	callgraphCmd.Flags().IntVarP(&callgraphWorkerNumber, "worker", "w", 8, "Number of workers")
	callgraphCmd.Flags().StringVarP(&callgraphCSVFolder, "csvoutput", "o", "", "Output folder for csv files")
	callgraphCmd.Flags().BoolVarP(&callgraphInitSchema, "init", "i", false, "Whether to init db schema ")
	callgraphCmd.Flags().BoolVarP(&callgraphDynamicExports, "dynamic", "d", true, "Whether to use dynamic export merge logic")
}
