package cmd

import (
	"database/sql"
	"encoding/csv"
	"github.com/markuszm/npm-analysis/codeanalysis/packagecallgraph"
	"github.com/markuszm/npm-analysis/database"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var (
	packageCostMysqlUrl     string
	packageCostPackageName  string
	packageCostFile         string
	packageCostOutput       string
	packageCostWorkerNumber int
)

var costSyncMap = sync.Map{}

var packageCostDB *sql.DB

var packageCostCmd = &cobra.Command{
	Use:   "packageCost",
	Short: "Calculates package cost",
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		mysqlInitializer := &database.Mysql{}
		var err error
		packageCostDB, err = mysqlInitializer.InitDB(packageCostMysqlUrl)
		if err != nil {
			logger.Fatal(err)
		}
		defer packageCostDB.Close()

		var packages []string

		if packageCostFile != "" {
			logger.Infow("using package input file", "file", packageCostFile)
			file, err := ioutil.ReadFile(packageCostFile)
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

		if packageCostPackageName != "" {
			packages = append(packages, packageCostPackageName)
		}

		costWorkerWait := sync.WaitGroup{}
		csvWorkerWait := sync.WaitGroup{}

		jobs := make(chan string, 10)
		csvChan := make(chan []string, 10)

		csvWorkerWait.Add(1)
		go csvWorkerPackageCost(csvChan, &csvWorkerWait)

		for w := 1; w <= packageCostWorkerNumber; w++ {
			costWorkerWait.Add(1)
			go costWorker(w, jobs, csvChan, &costWorkerWait)
		}

		for _, p := range packages {
			jobs <- p
		}

		close(jobs)

		logger.Info("closed jobs channel")

		costWorkerWait.Wait()

		close(csvChan)
		logger.Info("closed csv channels")

		csvWorkerWait.Wait()

		combinedCount := 0
		costSyncMap.Range(func(key, value interface{}) bool {
			if value.(bool) {
				combinedCount++
			}
			return true
		})

		logger.Infof("Combined package cost is %v", combinedCount)
	},
}

func csvWorkerPackageCost(csvChan chan []string, waitGroup *sync.WaitGroup) {
	file, err := os.Create(path.Join(packageCostOutput, "packageCost.csv"))
	if err != nil {
		logger.Fatal("cannot create result file")
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)

	for r := range csvChan {
		err := csvWriter.Write(r)
		if err != nil {
			logger.Fatal("cannot write result to csv")
		}
		csvWriter.Flush()
	}

	waitGroup.Done()
}

func costWorker(id int, jobs chan string, csvChan chan []string, waitGroup *sync.WaitGroup) {
	combinedPackageCost := make(map[string]bool, 0)

	for p := range jobs {
		packageCostIndependent := make(map[string]bool, 0)
		packagecallgraph.PackageCost(p, packageCostIndependent, packageCostDB)
		count := 0
		for _, ok := range packageCostIndependent {
			if ok {
				count++
			}
		}

		result := []string{p, strconv.Itoa(count)}
		csvChan <- result

		packagecallgraph.PackageCost(p, combinedPackageCost, packageCostDB)

		logger.Infow("Finished", "worker", id, "package", p)
	}

	for p, ok := range combinedPackageCost {
		if ok {
			costSyncMap.Store(p, ok)
		}
	}

	waitGroup.Done()
}

func init() {
	rootCmd.AddCommand(packageCostCmd)

	packageCostCmd.Flags().StringVarP(&packageCostMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	packageCostCmd.Flags().StringVarP(&packageCostPackageName, "package", "p", "", "package name")
	packageCostCmd.Flags().StringVarP(&packageCostFile, "file", "f", "", "file name to load package names from")
	packageCostCmd.Flags().StringVarP(&packageCostOutput, "output", "o", "./output/packageCost", "output folder")
	packageCostCmd.Flags().IntVarP(&packageCostWorkerNumber, "worker", "w", 20, "number of workers")
}
