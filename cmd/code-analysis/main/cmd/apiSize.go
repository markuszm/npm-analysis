package cmd

import (
	"database/sql"
	"encoding/json"
	"github.com/markuszm/npm-analysis/codeanalysis/packagecallgraph"
	"github.com/markuszm/npm-analysis/database"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var apiSizeNeo4jUrl string
var apiSizeMysqlUrl string
var apiSizeOutput string
var apiSizeInputFile string
var apiSizeWorkers int
var apiSizeHeuristic bool

var apiSizeMysql *sql.DB

var apiSizeCmd = &cobra.Command{
	Use:   "apiSize",
	Short: "Generates api size statistic",
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		var err error
		mysqlInitializer := &database.Mysql{}
		apiSizeMysql, err = mysqlInitializer.InitDB(callgraphMysqlUrl)
		if err != nil {
			logger.Fatal(err)
		}
		defer apiSizeMysql.Close()

		packageChan := make(chan string, 0)
		resultsChan := make(chan ApiSizeStats, 0)

		if apiSizeHeuristic {
			logger.Info("Using heuristic to calculate exported functions")
		}

		if apiSizeInputFile == "" {
			queries, err := packagecallgraph.NewGraphQueries(apiSizeNeo4jUrl)
			if err != nil {
				log.Fatal(err)
			}
			defer queries.Close()

			go queries.StreamPackages(packageChan)
		} else {
			go streamPackageNamesFromFileApiSize(packageChan)
		}

		writerWait := sync.WaitGroup{}
		workerWait := sync.WaitGroup{}
		writerWait.Add(1)
		go writeResultsApiSize(resultsChan, &writerWait)

		for w := 1; w <= apiSizeWorkers; w++ {
			workerWait.Add(1)
			go apiSizeCalculator(w, packageChan, resultsChan, &workerWait)
		}

		workerWait.Wait()
		close(resultsChan)

		writerWait.Wait()
	},
}

func apiSizeCalculator(workerId int, packageChan chan string, resultsChan chan ApiSizeStats, workerWait *sync.WaitGroup) {
	queries, err := packagecallgraph.NewGraphQueries(apiSizeNeo4jUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer queries.Close()

	for pkg := range packageChan {
		mainModuleName := getMainModuleName(apiSizeMysql, pkg)
		exportedFunctions, err := queries.GetExportedFunctionsForPackageActual(pkg, mainModuleName)
		if err != nil {
			log.Fatal(err)
		}

		if len(exportedFunctions) == 0 {
			if apiSizeHeuristic {
				exportedFunctions, err = queries.GetExportedFunctionsForPackageHeuristic(pkg, mainModuleName)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				continue
			}
		}

		numberOfExportedFunctions := len(exportedFunctions)

		apiSizeStat := ApiSizeStats{
			PackageName:       pkg,
			ExportedFunctions: exportedFunctions,
			Size:              numberOfExportedFunctions,
		}
		resultsChan <- apiSizeStat

		logger.Infof("Worker %v finished with package %s", workerId, pkg)
	}

	workerWait.Done()
}

func writeResultsApiSize(resultStream chan ApiSizeStats, waitGroup *sync.WaitGroup) {
	file, err := os.Create(apiSizeOutput)
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

func streamPackageNamesFromFileApiSize(packageChan chan string) {
	logger.Infow("using package input file", "file", apiSizeInputFile)
	file, err := ioutil.ReadFile(apiSizeInputFile)
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
	rootCmd.AddCommand(apiSizeCmd)

	apiSizeCmd.Flags().StringVarP(&apiSizeNeo4jUrl, "neo4j", "n", "bolt://neo4j:npm@localhost:7689", "Neo4j bolt url")
	apiSizeCmd.Flags().StringVarP(&apiSizeMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	apiSizeCmd.Flags().StringVarP(&apiSizeOutput, "output", "o", "/home/markus/npm-analysis/apiSize.json", "output file")
	apiSizeCmd.Flags().StringVarP(&apiSizeInputFile, "input", "i", "", "input file containing list with package names")
	apiSizeCmd.Flags().IntVarP(&apiSizeWorkers, "worker", "w", 20, "number of workers")
	apiSizeCmd.Flags().BoolVar(&apiSizeHeuristic, "heuristic", true, "use heuristic for packages that have no actual exports")
}

type ApiSizeStats struct {
	PackageName       string
	ExportedFunctions []string
	Size              int
}
