package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var exportUsageNeo4jUrl string
var exportUsageMysqlUrl string
var exportUsageOutput string
var exportUsageInputFile string

var exportUsageCmd = &cobra.Command{
	Use:   "exportUsage",
	Short: "Generates export usage stats for packages",
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		queries, err := packagecallgraph.NewGraphQueries(exportUsageNeo4jUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer queries.Close()

		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(callgraphMysqlUrl)
		if databaseInitErr != nil {
			logger.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		packageChan := make(chan string, 0)

		if exportUsageInputFile == "" {
			go queries.StreamPackages(packageChan)
		} else {
			go streamPackageNamesFromFile(packageChan)
		}

		count := 0

		file, err := os.Create(exportUsageOutput)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)

		for pkg := range packageChan {
			requiredPackages, err := queries.GetRequiredPackagesForPackage(pkg)
			if err != nil {
				log.Fatal(err)
			}

			functionsUsedMap := make(map[string][]string, 0)
			percentageUsedMap := make(map[string]float64, 0)

			for _, requiredPkg := range requiredPackages {
				mainModuleName := getMainModuleName(mysql, requiredPkg)
				exportedFunctions, err := queries.GetExportedFunctionsForPackage(requiredPkg, mainModuleName)
				if err != nil {
					log.Fatal(err)
				}

				var usedFunctions []string
				for _, function := range exportedFunctions {
					packageFunctions, err := queries.GetFunctionsFromPackageThatCallAnotherFunctionDirectly(pkg, function)
					if err != nil {
						log.Fatal(err)
					}
					if len(packageFunctions) > 0 {
						usedFunctions = append(usedFunctions, function)
					}

				}

				numberOfExportedFunctions := len(exportedFunctions)
				numberOfUsedFunctions := len(usedFunctions)

				functionsUsedMap[requiredPkg] = usedFunctions
				percentageUsedMap[requiredPkg] = float64(numberOfUsedFunctions) / float64(numberOfExportedFunctions)
			}

			exportUsageStat := ExportUsageStats{
				PackageName:       pkg,
				PackagesUsed:      requiredPackages,
				FunctionsUsed:     functionsUsedMap,
				PercentageUsed:    percentageUsedMap,
				PackagesUsedCount: len(requiredPackages),
			}

			err = encoder.Encode(exportUsageStat)
			if err != nil {
				log.Fatal("Cannot write to result file", err)
			}

			count++

			if count%1000 == 0 {
				logger.Infof("Finished %v packages", count)
			}
		}

	},
}

func getMainModuleName(mysql *sql.DB, packageName string) string {
	mainFile, err := database.MainFileForPackage(mysql, packageName)
	if err != nil {
		logger.Fatalf("error getting mainFile from database for moduleName %s with error %s", packageName, err)
	}
	// cleanup main file
	mainFile = strings.TrimSuffix(mainFile, filepath.Ext(mainFile))
	mainFile = strings.TrimLeft(mainFile, "./")
	if mainFile == "" {
		mainFile = "index"
	}

	return fmt.Sprintf("%s|%s", packageName, mainFile)
}

func streamPackageNamesFromFile(packageChan chan string) {
	logger.Infow("using package input file", "file", exportUsageInputFile)
	file, err := ioutil.ReadFile(exportUsageInputFile)
	if err != nil {
		logger.Fatalw("could not read file", "err", err)
	}
	lines := strings.Split(string(file), "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		packageChan <- l
	}

	close(packageChan)
}

func init() {
	rootCmd.AddCommand(exportUsageCmd)

	exportUsageCmd.Flags().StringVarP(&exportUsageNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7689", "Neo4j bolt url")
	exportUsageCmd.Flags().StringVarP(&exportUsageMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	exportUsageCmd.Flags().StringVarP(&exportUsageOutput, "output", "o", "/home/markus/npm-analysis/exportUsage.json", "output file")
	exportUsageCmd.Flags().StringVarP(&exportUsageInputFile, "input", "i", "", "input file containing list with package names")
}

type ExportUsageStats struct {
	PackageName       string
	PackagesUsed      []string
	FunctionsUsed     map[string][]string
	PercentageUsed    map[string]float64
	PackagesUsedCount int
}
