package cmd

import (
	"github.com/markuszm/npm-analysis/codeanalysis/resultprocessing"
	"github.com/spf13/cobra"
	"path"
)

var fileDistributionInput string
var fileDistributionOutput string

// fileDistributionCmd represents the fileDistribution command
var fileDistributionCmd = &cobra.Command{
	Use:   "fileDistribution",
	Short: "Result processing for file distribution analysis results",
	Long:  `Processes analysis results to create plotable csv results`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		allPackages, err := resultprocessing.MergeFileDistributionResult(fileDistributionInput, 0)
		if err != nil {
			logger.Fatal(err)
		}
		resultprocessing.WriteFiledistributionResult(allPackages, path.Join(fileDistributionOutput, "allpackages.csv"))

		percentages, err := resultprocessing.CalculatePercentageForEachPackage(fileDistributionInput)
		if err != nil {
			logger.Fatal(err)
		}

		err = resultprocessing.WritePercentagesPerPackageForExtension(percentages, path.Join(fileDistributionOutput, "percentages.csv"))
		if err != nil {
			logger.Fatal(err)
		}

		err = resultprocessing.WriteCumulativeBinaryGraph(fileDistributionInput, path.Join(fileDistributionOutput, "binaryCount.json"))
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(fileDistributionCmd)

	fileDistributionCmd.Flags().StringVarP(&fileDistributionInput, "input", "i", "./output/filedistribution.json", "path to file containing analysis results")
	fileDistributionCmd.Flags().StringVarP(&fileDistributionOutput, "output", "o", "./output/filedistribution", "output path")
}
