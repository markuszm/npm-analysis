package cmd

import (
	"github.com/markuszm/npm-analysis/codeanalysispipeline/resultprocessing"
	"github.com/spf13/cobra"
)

var inputPath string
var outputPath string
var analysis string

// resultCmd represents the result command
var resultCmd = &cobra.Command{
	Use:   "result",
	Short: "Result processing for analysis results",
	Long:  `Processes analysis results e.g. plots them or creates other insights`,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := resultprocessing.MergeFileDistributionResult(inputPath, 1000)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Info(result)
	},
}

func init() {
	rootCmd.AddCommand(resultCmd)

	resultCmd.Flags().StringVarP(&inputPath, "input", "i", "/home/markus/npm-analysis/filedistribution.json", "path to file containing analysis results")
	resultCmd.Flags().StringVarP(&outputPath, "output", "o", "/home/markus/npm-analysis/filedistribution", "output path")
	resultCmd.Flags().StringVarP(&analysis, "analysis", "a", "file_distribution", "specify for which analysis to process results")
}
