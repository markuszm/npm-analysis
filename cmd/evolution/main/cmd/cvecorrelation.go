package cmd

import (
	"encoding/json"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
)

const cveCorrelationLatestDependenciesJsonPath = "./db-data/latestDependenciesTimeline.json"

var cveCorrelationCVEInput string

var cveCorrelationOutputFolder string

var cveCorrelationLatestDependencies map[string]map[string]bool

// calculates Package reach of a packages and plots it
var cveCorrelationCmd = &cobra.Command{
	Use:   "cveCorrelate",
	Short: "calculates correlation between vulnerable and transitive dependencies",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		cveCorrelationLatestDependencies = reach.LoadJSONLatestDependencies(cveCorrelationLatestDependenciesJsonPath)
		calculateCorrelation()
	},
}

func calculateCorrelation() {
	file, err := ioutil.ReadFile(cveCorrelationCVEInput)
	if err != nil {
		logger.Fatal(err)
	}

	var cvePackageList []string
	err = json.Unmarshal(file, &cvePackageList)
	if err != nil {
		logger.Fatal(err)
	}

	var cvePackageSet = make(map[string]bool, 0)

	for _, p := range cvePackageList {
		cvePackageSet[p] = true
	}

	var results []CveCorrelationResult

	for p, deps := range cveCorrelationLatestDependencies {
		result := CveCorrelationResult{
			PackageName:       p,
			PackageCost:       len(deps),
			IsVulnerable:      false,
			VulnerabilityCost: 0,
		}
		vulnCost := 0
		for d, ok := range deps {
			if ok && cvePackageSet[d] {
				result.IsVulnerable = true
				vulnCost++
			}
		}

		result.VulnerabilityCost = vulnCost
		results = append(results, result)
		logger.Debugf("Finished %s", p)
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath := path.Join(cveCorrelationOutputFolder, "cveCorrelation.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
}

type CveCorrelationResult struct {
	PackageName       string
	IsVulnerable      bool
	VulnerabilityCost int
	PackageCost       int
}

func init() {
	rootCmd.AddCommand(cveCorrelationCmd)

	cveCorrelationCmd.Flags().StringVar(&cveCorrelationCVEInput, "cveInput", "./cvePackages.json", "input file containing packages that are vulnerable")
	cveCorrelationCmd.Flags().StringVar(&cveCorrelationOutputFolder, "output", "./output/cveCorrelation", "output folder for results")

}
