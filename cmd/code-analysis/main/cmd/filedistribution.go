package cmd

import (
	"github.com/markuszm/npm-analysis/codeanalysispipeline/resultprocessing"
	"github.com/spf13/cobra"
	"path"
)

var inputPath string
var outputPath string

// resultCmd represents the result command
var resultCmd = &cobra.Command{
	Use:   "filedistribution",
	Short: "Result processing for filedistribution results",
	Long:  `Processes analysis results e.g. plots them or creates other insights`,
	Run: func(cmd *cobra.Command, args []string) {
		allPackages, err := resultprocessing.MergeFileDistributionResult(inputPath, 0)
		if err != nil {
			logger.Fatal(err)
		}
		resultprocessing.WriteFiledistributionResult(allPackages, path.Join(outputPath, "allpackages.csv"))

		percentages, err := resultprocessing.CalculatePercentageForEachPackage(inputPath)
		if err != nil {
			logger.Fatal(err)
		}

		err = resultprocessing.WritePercentagesPerPackageForExtension(percentages, path.Join(outputPath, "percentages.csv"))
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(resultCmd)

	resultCmd.Flags().StringVarP(&inputPath, "input", "i", "/home/markus/npm-analysis/filedistribution.json", "path to file containing analysis results")
	resultCmd.Flags().StringVarP(&outputPath, "output", "o", "/home/markus/npm-analysis/filedistribution", "output path")
}
