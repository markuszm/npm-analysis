package cmd

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"github.com/markuszm/npm-analysis/codeanalysis/packagecallgraph"
	"github.com/markuszm/npm-analysis/database"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
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
	packageReachLayered      bool
	packageReachDev          bool
	packageReachCombined     bool
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

		if packageReachLayered {
			packageReachWorkerNumber = 1
		}

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

		if packageReachCombined {
			combinedCount := 0
			reachSyncMap.Range(func(key, value interface{}) bool {
				if value.(bool) {
					combinedCount++
				}
				return true
			})

			logger.Infof("Combined package reach is %v", combinedCount)
		}

	},
}

func csvWorker(csvChan chan []string, waitGroup *sync.WaitGroup) {
	fileName := "packageReach.csv"
	file, err := os.Create(path.Join(packageReachOutput, fileName))
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

	// only for layered calculation
	combinedPackageReachLayered := make(map[string]packagecallgraph.ReachDetails, 0)
	layersCountMap := make(map[int]int, 0)
	alreadyFoundDependents := make(map[string]bool, 0)

	for p := range jobs {
		if packageReachLayered {
			packagecallgraph.PackageReachLayer(p, combinedPackageReachLayered, db, 1)
			for dependent, details := range combinedPackageReachLayered {
				if !alreadyFoundDependents[dependent] {
					layersCountMap[details.Layer] = layersCountMap[details.Layer] + 1
					alreadyFoundDependents[dependent] = true
				}
			}
		} else {
			packagesReachedIndependent := make(map[string]bool, 0)

			if packageReachDev {
				packagecallgraph.PackageReachDev(p, packagesReachedIndependent, db)
			} else {
				packagecallgraph.PackageReach(p, packagesReachedIndependent, db)
			}

			count := 0
			for _, ok := range packagesReachedIndependent {
				if ok {
					count++
				}
			}

			result := []string{p, strconv.Itoa(count)}
			csvChan <- result
		}

		if packageReachCombined {
			packagecallgraph.PackageReach(p, combinedPackageReach, db)
		}

		logger.Infow("Finished", "worker", id, "package", p)
	}

	if packageReachLayered {
		fileName := "packageReachLayered.json"

		bytes, err := json.Marshal(layersCountMap)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path.Join(packageReachOutput, fileName), bytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if packageReachCombined {
		for p, ok := range combinedPackageReach {
			if ok {
				reachSyncMap.Store(p, ok)
			}
		}
	}

	waitGroup.Done()
}

func init() {
	rootCmd.AddCommand(packageReachCmd)

	packageReachCmd.Flags().StringVarP(&packageReachMysqlUrl, "mysql", "m", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "mysql url")
	packageReachCmd.Flags().StringVarP(&packageReachPackageName, "package", "p", "", "package name")
	packageReachCmd.Flags().StringVarP(&packageReachFile, "file", "f", "", "file name to load package names from")
	packageReachCmd.Flags().StringVarP(&packageReachOutput, "output", "o", "./output/packageReach", "output folder")
	packageReachCmd.Flags().IntVarP(&packageReachWorkerNumber, "worker", "w", 20, "number of workers")
	packageReachCmd.Flags().BoolVarP(&packageReachLayered, "layers", "l", false, "whether to calculate layered result")
	packageReachCmd.Flags().BoolVarP(&packageReachDev, "dev", "d", false, "whether to include dev dependents")
	packageReachCmd.Flags().BoolVarP(&packageReachCombined, "combined", "c", false, "whether to calculate combined package reach")
}
