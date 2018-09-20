package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	neoQueryGeneratorChainLength            int
	neoQueryGeneratorInitialFunctionBuiltIn bool
	neoQueryGeneratorInitialFunctionName    string
	neoQueryGeneratorQueryType              string
)

var neoQueryGeneratorCmd = &cobra.Command{
	Use:   "neoQueryGenerator",
	Short: "Generates neo4j package callgraph queries",
	Run: func(cmd *cobra.Command, args []string) {
		switch neoQueryGeneratorQueryType {
		case "package-callchain":
			query := ""
			chainMatch := "MATCH "
			whereClause := "WHERE "
			returnClause := "RETURN "
			for i := 1; i <= neoQueryGeneratorChainLength; i++ {
				if i == 1 {
					chainMatch += fmt.Sprintf("(f%v:Function {name: \"%s\"})", i, neoQueryGeneratorInitialFunctionName)
					if !neoQueryGeneratorInitialFunctionBuiltIn {
						whereClause += fmt.Sprintf("p%v.name ", i)
					}
					returnClause += fmt.Sprintf("f%v", i)
				} else {
					chainMatch += fmt.Sprintf("<-[:CALL]-(f%v:Function)", i)
					if i != 2 && !neoQueryGeneratorInitialFunctionBuiltIn {
						whereClause += fmt.Sprintf("<> p%v.name ", i)
					}
					returnClause += fmt.Sprintf(",f%v", i)
				}

				if i == 1 && neoQueryGeneratorInitialFunctionBuiltIn {
					continue
				}
				query += fmt.Sprintf(",(f%v)<-[:CONTAINS_FUNCTION]-(:Module)<-[:CONTAINS_MODULE]-(p%v:Package)", i, i)
			}

			if whereClause == "WHERE " {
				whereClause = ""
			}

			fullQuery := fmt.Sprintf("%s%s %s %s", chainMatch, query, whereClause, returnClause)
			println(fullQuery)
		}

	},
}

func init() {
	rootCmd.AddCommand(neoQueryGeneratorCmd)

	neoQueryGeneratorCmd.Flags().IntVarP(&neoQueryGeneratorChainLength, "length", "l", 3, "length of call chain")
	neoQueryGeneratorCmd.Flags().BoolVarP(&neoQueryGeneratorInitialFunctionBuiltIn, "builtin", "b", false, "whether initial function is built in ")
	neoQueryGeneratorCmd.Flags().StringVarP(&neoQueryGeneratorInitialFunctionName, "function", "f", "eval", "name of first function to match for")
	neoQueryGeneratorCmd.Flags().StringVarP(&neoQueryGeneratorQueryType, "query", "q", "package-callchain", "type of query")
}
