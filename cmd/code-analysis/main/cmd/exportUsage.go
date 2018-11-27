package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis/packagecallgraph"
	"github.com/markuszm/npm-analysis/database"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var exportUsageNeo4jUrl string
var exportUsageMysqlUrl string
var exportUsageOutput string
var exportUsageInputFile string
var exportUsageWorkers int
var exportUsageHeuristic bool

var mysql *sql.DB

var exportUsageCmd = &cobra.Command{
	Use:   "exportUsage",
	Short: "Generates export usage stats for packages",
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		var err error
		mysqlInitializer := &database.Mysql{}
		mysql, err = mysqlInitializer.InitDB(callgraphMysqlUrl)
		if err != nil {
			logger.Fatal(err)
		}
		defer mysql.Close()

		packageChan := make(chan string, 0)
		resultsChan := make(chan ExportUsageStats, 0)

		if exportUsageHeuristic {
			logger.Info("Using heuristic to calculate exported functions")
		}

		if exportUsageInputFile == "" {
			queries, err := packagecallgraph.NewGraphQueries(exportUsageNeo4jUrl)
			if err != nil {
				log.Fatal(err)
			}
			defer queries.Close()

			go queries.StreamPackages(packageChan)
		} else {
			go streamPackageNamesFromFile(packageChan)
		}

		writerWait := sync.WaitGroup{}
		workerWait := sync.WaitGroup{}
		writerWait.Add(1)
		go writeResults(resultsChan, &writerWait)

		for w := 1; w <= exportUsageWorkers; w++ {
			workerWait.Add(1)
			go exportUsageCalculator(w, packageChan, resultsChan, &workerWait)
		}

		workerWait.Wait()
		close(resultsChan)

		writerWait.Wait()
	},
}

func exportUsageCalculator(workerId int, packageChan chan string, resultsChan chan ExportUsageStats, workerWait *sync.WaitGroup) {
	queries, err := packagecallgraph.NewGraphQueries(exportUsageNeo4jUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer queries.Close()

	for pkg := range packageChan {
		requiredPackages, err := queries.GetRequiredPackagesForPackage(pkg)
		if err != nil {
			log.Fatal(err)
		}

		functionsUsedMap := make(map[string][]string, 0)
		percentageUsedMap := make(map[string]float64, 0)

		for _, requiredPkg := range requiredPackages {
			mainModuleName := getMainModuleName(mysql, requiredPkg)
			exportedFunctions, err := queries.GetExportedFunctionsForPackageActual(requiredPkg, mainModuleName)
			if err != nil {
				log.Fatal(err)
			}

			if len(exportedFunctions) == 0 {
				if exportUsageHeuristic {
					exportedFunctions, err = queries.GetExportedFunctionsForPackageHeuristic(requiredPkg, mainModuleName)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					continue
				}
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

			percentage := float64(numberOfUsedFunctions) / float64(numberOfExportedFunctions)
			if numberOfExportedFunctions == 0 {
				percentage = -1.0
				logger.Warnw("exported functions is zero", "package", requiredPkg)
			}

			functionsUsedMap[requiredPkg] = usedFunctions
			percentageUsedMap[requiredPkg] = percentage
		}

		exportUsageStat := ExportUsageStats{
			PackageName:       pkg,
			PackagesUsed:      requiredPackages,
			FunctionsUsed:     functionsUsedMap,
			PercentageUsed:    percentageUsedMap,
			PackagesUsedCount: len(requiredPackages),
		}
		resultsChan <- exportUsageStat

		logger.Infof("Worker %v finished with package %s", workerId, pkg)
	}

	workerWait.Done()
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

func writeResults(resultStream chan ExportUsageStats, waitGroup *sync.WaitGroup) {
	file, err := os.Create(exportUsageOutput)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	for r := range resultStream {
		err = encoder.Encode(r)
		if err != nil {
			log.Fatal("Cannot write to result file ", err)
		}
	}

	waitGroup.Done()
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
	exportUsageCmd.Flags().IntVarP(&exportUsageWorkers, "worker", "w", 20, "number of workers")
	exportUsageCmd.Flags().BoolVar(&exportUsageHeuristic, "heuristic", true, "use heuristic for packages that have no actual exports")
}

type ExportUsageStats struct {
	PackageName       string
	PackagesUsed      []string
	FunctionsUsed     map[string][]string
	PercentageUsed    map[string]float64
	PackagesUsedCount int
}
