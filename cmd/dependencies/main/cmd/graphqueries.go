package cmd

import (
	"encoding/csv"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var graphQueriesNeo4jUrl string

var graphQueriesOutputFile string

func init() {
	rootCmd.AddCommand(graphQueriesCmd)

	graphQueriesCmd.Flags().StringVar(&graphQueriesOutputFile, "output", "/home/markus/npm-analysis/packages_with_dependents", "result file")
	graphQueriesCmd.Flags().StringVar(&graphQueriesNeo4jUrl, "neo4j", "bolt://neo4j:npm@localhost:7687", "full connection url to neo4j server")
}

var graphQueriesCmd = &cobra.Command{
	Use:   "graphQueries",
	Short: "Query the package callgraph",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		packageChan := make(chan []string, 0)

		go queryPackagesWithDependents(packageChan)

		file, err := os.Create(graphQueriesOutputFile)
		if err != nil {
			log.Fatal(err)
		}

		writer := csv.NewWriter(file)

		for p := range packageChan {
			log.Print(p)
			err := writer.Write(p)
			if err != nil {
				log.Fatal(err)
			}
		}

		writer.Flush()
	},
}

func queryPackagesWithDependents(packageChan chan []string) {
	database := graph.NewNeo4JDatabase()
	defer database.Close()
	err := database.InitDB(graphQueriesNeo4jUrl)
	if err != nil {
		log.Fatal(err)
	}
	resultChan := make(chan []interface{}, 0)
	go database.QueryStream("MATCH (p:Package)<-[:DEPEND|:DEPEND_DEV|:DEPEND_PEER|:DEPEND_OPTIONAL]-(d:Package) RETURN DISTINCT p.name, p.version", map[string]interface{}{}, resultChan)
	for r := range resultChan {
		// TODO: very unsafe - should check if result is valid
		if r[1] == nil {
			continue
		}
		packageChan <- []string{r[0].(string), r[1].(string)}
	}
	close(packageChan)
}
