package cmd

import (
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/spf13/cobra"
	"path"
)

var usedDependenciesInput string
var usedDependenciesOutput string

// usedDependenciesCmd represents the usedDependencies command
var usedDependenciesCmd = &cobra.Command{
	Use:   "usedDependencies",
	Short: "Creates csv statistics for used dependency analysis results",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		dependenciesRatios, err := resultprocessing.CalculateUsedDependenciesRatio(usedDependenciesInput)
		if err != nil {
			logger.Fatal(err)
		}

		err = resultprocessing.WriteUsedDependencyRatios(dependenciesRatios, path.Join(usedDependenciesOutput, "ratios.csv"))
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(usedDependenciesCmd)

	usedDependenciesCmd.Flags().StringVarP(&usedDependenciesInput, "input", "i", "/home/markus/npm-analysis/usedDependencies.json", "path to file containing analysis results")
	usedDependenciesCmd.Flags().StringVarP(&usedDependenciesOutput, "output", "o", "/home/markus/npm-analysis/usedDependencies", "output path")
}
