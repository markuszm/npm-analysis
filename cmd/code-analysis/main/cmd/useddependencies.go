package cmd

import (
	"github.com/markuszm/npm-analysis/codeanalysispipeline/resultprocessing"
	"github.com/spf13/cobra"
	"path"
)

// usedDependenciesCmd represents the usedDependencies command
var usedDependenciesCmd = &cobra.Command{
	Use:   "usedDependencies",
	Short: "Creates csv statistics for used dependency analysis results",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		dependenciesRatios, err := resultprocessing.CalculateUsedDependenciesRatio(inputPath)
		if err != nil {
			logger.Fatal(err)
		}

		err = resultprocessing.WriteUsedDependencyRatios(dependenciesRatios, path.Join(outputPath, "ratios.csv"))
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(usedDependenciesCmd)

	usedDependenciesCmd.Flags().StringVarP(&inputPath, "input", "i", "/home/markus/npm-analysis/usedDependencies.json", "path to file containing analysis results")
	usedDependenciesCmd.Flags().StringVarP(&outputPath, "output", "o", "/home/markus/npm-analysis/usedDependencies", "output path")
}
