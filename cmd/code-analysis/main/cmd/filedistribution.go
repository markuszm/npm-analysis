package cmd

import (
	"github.com/markuszm/npm-analysis/codeanalysispipeline/resultprocessing"
	"github.com/spf13/cobra"
	"path"
)

// fileDistributionCmd represents the fileDistribution command
var fileDistributionCmd = &cobra.Command{
	Use:   "fileDistribution",
	Short: "Result processing for file distribution analysis results",
	Long:  `Processes analysis results to create plotable csv results`,
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
	rootCmd.AddCommand(fileDistributionCmd)

	fileDistributionCmd.Flags().StringVarP(&inputPath, "input", "i", "/home/markus/npm-analysis/filedistribution.json", "path to file containing analysis results")
	fileDistributionCmd.Flags().StringVarP(&outputPath, "output", "o", "/home/markus/npm-analysis/filedistribution", "output path")
}
