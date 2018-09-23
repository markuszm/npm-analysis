package cmd

import (
	"database/sql"
	"encoding/csv"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/packagecallgraph"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var (
	packageReachMysqlUrl     string
	packageReachPackageName  string
	packageReachFile         string
	packageReachOutput       string
	packageReachWorkerNumber int
)

var reachSyncMap = sync.Map{}

var db *sql.DB

// callgraphCmd represents the callgraph command
var packageReachCmd = &cobra.Command{
	Use:   "packageReach",
	Short: "Calculates package reach",
	Long:  `Calculates package reach for a given package`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		mysqlInitializer := &database.Mysql{}
		var err error
		db, err = mysqlInitializer.InitDB(packageReachMysqlUrl)
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

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

		reachWorkerWait := sync.WaitGroup{}
		csvWorkerWait := sync.WaitGroup{}

		jobs := make(chan string, 10)
		csvChan := make(chan []string, 10)

		csvWorkerWait.Add(1)
		go csvWorker(csvChan, &csvWorkerWait)

		for w := 1; w <= packageReachWorkerNumber; w++ {
			reachWorkerWait.Add(1)
			go reachWorker(w, jobs, csvChan, &reachWorkerWait)
		}

		for _, p := range packages {
			jobs <- p
		}

		close(jobs)

		logger.Info("closed jobs channel")

		reachWorkerWait.Wait()

		close(csvChan)
		logger.Info("closed csv channels")

		csvWorkerWait.Wait()

		combinedCount := 0
		reachSyncMap.Range(func(key, value interface{}) bool {
			if value.(bool) {
				combinedCount++
			}
			return true
		})

		logger.Infof("Combined package reach is %v", combinedCount)
	},
}

func csvWorker(csvChan chan []string, waitGroup *sync.WaitGroup) {
	file, err := os.Create(path.Join(packageReachOutput, "packagesReach.csv"))
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

func reachWorker(id int, jobs chan string, csvChan chan []string, waitGroup *sync.WaitGroup) {
	combinedPackageReach := make(map[string]bool, 0)

	for p := range jobs {
		packagesReachedIndependent := make(map[string]bool, 0)
		packagecallgraph.PackageReach(p, packagesReachedIndependent, db)
		count := 0
		for _, ok := range packagesReachedIndependent {
			if ok {
				count++
			}
		}

		result := []string{p, strconv.Itoa(count)}
		csvChan <- result

		packagecallgraph.PackageReach(p, combinedPackageReach, db)

		logger.Infow("Finished", "worker", id, "package", p)
	}

	for p, ok := range combinedPackageReach {
		if ok {
			reachSyncMap.Store(p, ok)
		}
	}

	waitGroup.Done()
}

func init() {
	rootCmd.AddCommand(packageReachCmd)

	packageReachCmd.Flags().StringVarP(&packageReachMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	packageReachCmd.Flags().StringVarP(&packageReachPackageName, "package", "p", "", "package name")
	packageReachCmd.Flags().StringVarP(&packageReachFile, "file", "f", "", "file name to load package names from")
	packageReachCmd.Flags().StringVarP(&packageReachOutput, "output", "o", "/home/markus/npm-analysis", "output folder")
	packageReachCmd.Flags().IntVarP(&packageReachWorkerNumber, "worker", "w", 20, "number of workers")
}
