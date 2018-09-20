package cmd

import (
	"encoding/csv"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	packageReachMysqlUrl    string
	packageReachPackageName string
	packageReachFile        string
	packageReachOutput      string
)

// callgraphCmd represents the callgraph command
var packageReachCmd = &cobra.Command{
	Use:   "packageReach",
	Short: "Calculates package reach",
	Long:  `Calculates package reach for a given package`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		mysqlInitializer := &database.Mysql{}
		mysql, err := mysqlInitializer.InitDB(packageReachMysqlUrl)
		if err != nil {
			logger.Fatal(err)
		}
		defer mysql.Close()

		var packages []string

		if packageReachFile != "" {
			logger.Infow("using package input file", "file", packageReachFile)
			file, err := ioutil.ReadFile(packageReachFile)
			if err != nil {
				logger.Fatalw("could not read file", "err", err)
			}
			lines := strings.Split(string(file), "\n")

			for _, l := range lines {
				if l == "" {
					continue
				}
				packages = append(packages, l)
			}
		}

		if packageReachPackageName != "" {
			packages = append(packages, packageReachPackageName)
		}

		file, err := os.Create(path.Join(packageReachOutput, "packagesReach.csv"))
		if err != nil {
			logger.Fatal("cannot create result file")
		}

		csvWriter := csv.NewWriter(file)

		packagesReached := make(map[string]bool, 0)

		for _, p := range packages {
			packagesReachedIndependent := make(map[string]bool, 0)
			packagecallgraph.PackageReach(p, packagesReachedIndependent, mysql)
			count := 0
			for _, ok := range packagesReachedIndependent {
				if ok {
					count++
				}
			}
			err := csvWriter.Write([]string{p, strconv.Itoa(count)})
			if err != nil {
				logger.Fatal("cannot write result to csv")
			}

			packagecallgraph.PackageReach(p, packagesReached, mysql)
			logger.Infow("Finished", "package", p)
		}
		csvWriter.Flush()

		combinedCount := 0
		for _, ok := range packagesReached {
			if ok {
				combinedCount++
			}
		}

		logger.Infof("Combined package reach is %v", combinedCount)

	},
}

func init() {
	rootCmd.AddCommand(packageReachCmd)

	packageReachCmd.Flags().StringVarP(&packageReachMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	packageReachCmd.Flags().StringVarP(&packageReachPackageName, "package", "p", "", "package name")
	packageReachCmd.Flags().StringVarP(&packageReachFile, "file", "f", "", "file name to load package names from")
	packageReachCmd.Flags().StringVarP(&packageReachOutput, "output", "o", "/home/markus/npm-analysis", "output folder")
}
