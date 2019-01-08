package cmd

import (
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

var graphInsertNeo4jUrl string

var graphInsertMysqlUser string

var graphInsertMysqlPassword string

var graphInsertWorkerNumber int

var graphInsertInsertType string
var graphInsertDepType string

func init() {
	rootCmd.AddCommand(graphInsertCmd)

	graphInsertCmd.Flags().StringVar(&graphInsertInsertType, "insert", "author", "type to insert")
	graphInsertCmd.Flags().StringVar(&graphInsertDepType, "type", "dependencies", "specify which type of dependency to insert")
	graphInsertCmd.Flags().StringVar(&graphInsertNeo4jUrl, "neo4j", "bolt://neo4j:npm@localhost:7687", "neo4j url")
	graphInsertCmd.Flags().StringVar(&graphInsertMysqlUser, "mysqlUser", "root", "mysql user")
	graphInsertCmd.Flags().StringVar(&graphInsertMysqlPassword, "mysqlPassword", "npm-analysis", "mysql password")
	graphInsertCmd.Flags().IntVar(&graphInsertWorkerNumber, "workers", 100, "number of workers")
}

var graphInsertCmd = &cobra.Command{
	Use:   "graphInsert",
	Short: "Insert metadata into graph database",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", graphInsertMysqlUser, graphInsertMysqlPassword))
		if databaseInitErr != nil {
			log.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		initGraphDatabase()

		count := 0

		workerWait := sync.WaitGroup{}

		jobs := make(chan interface{}, 100)

		for w := 1; w <= graphInsertWorkerNumber; w++ {
			workerWait.Add(1)
			go graphInsertWorker(w, jobs, &workerWait)
		}

		var jobItems []interface{}
		var retrieveErr error

		switch graphInsertInsertType {
		case "dependencies":
			var dependencies []model.Dependency
			dependencies, retrieveErr = database.GetAllDependencies(mysql, graphInsertDepType)
			jobItems = transformDependenciesToInterfaceSlice(dependencies)
		case "author":
			var persons []database.Person
			persons, retrieveErr = database.GetAuthors(mysql)
			jobItems = transformPersonToInterfaceSlice(persons)
		case "maintainers":
			var persons []database.Person
			persons, retrieveErr = database.GetMaintainers(mysql)
			jobItems = transformPersonToInterfaceSlice(persons)
		case "package":
			var packages []model.PackageVersionPair
			packages, retrieveErr = database.GetPackagesWithVersion(mysql)
			jobItems = transformPackagesToInterfaceSlice(packages)
		}
		if retrieveErr != nil {
			log.Fatal(retrieveErr)
		}

		for _, pkg := range jobItems {
			jobs <- pkg
		}

		close(jobs)

		log.Println(count)

		workerWait.Wait()
	},
}

func transformPersonToInterfaceSlice(persons []database.Person) []interface{} {
	var result []interface{}
	for _, p := range persons {
		result = append(result, p)
	}
	return result
}

func transformPackagesToInterfaceSlice(packages []model.PackageVersionPair) []interface{} {
	var result []interface{}
	for _, p := range packages {
		result = append(result, p)
	}
	return result
}

func transformDependenciesToInterfaceSlice(dependencies []model.Dependency) []interface{} {
	var result []interface{}
	for _, p := range dependencies {
		result = append(result, p)
	}
	return result
}

func initGraphDatabase() error {
	neo4JDatabase := graph.NewNeo4JDatabase()
	err := neo4JDatabase.InitDB(graphInsertNeo4jUrl)
	if err != nil {
		return err
	}

	graph.Init(neo4JDatabase)

	neo4JDatabase.Close()

	return nil
}

func graphInsertWorker(workerId int, jobs chan interface{}, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(graphInsertNeo4jUrl)
	if initErr != nil {
		log.Fatal(initErr)
	}

	for j := range jobs {
		var insertErr error
		switch graphInsertInsertType {
		case "author":
			author := j.(database.Person)
			person := model.Person{
				Name:  author.Name,
				Email: author.Email,
				Url:   author.Url,
			}
			packageName := author.PackageName

			insertErr = graph.InsertAuthorRelation(neo4JDatabase, person, packageName)
		case "maintainers":
			maintainer := j.(database.Person)
			person := model.Person{
				Name:  maintainer.Name,
				Email: maintainer.Email,
				Url:   maintainer.Url,
			}
			packageName := maintainer.PackageName

			insertErr = graph.InsertMaintainerRelation(neo4JDatabase, person, packageName)
		case "package":
			packageVersionPair := j.(model.PackageVersionPair)
			insertErr = graph.InsertPackage(neo4JDatabase, packageVersionPair)
		case "dependencies":
			dependency := j.(model.Dependency)
			switch graphInsertDepType {
			case "dependencies":
				insertErr = graph.InsertDependency(neo4JDatabase, dependency)
			case "bundledDependencies":
				insertErr = graph.InsertBundledDependency(neo4JDatabase, dependency)
			case "devDependencies":
				insertErr = graph.InsertDevDependency(neo4JDatabase, dependency)
			case "optionalDependencies":
				insertErr = graph.InsertOptionalDependency(neo4JDatabase, dependency)
			case "peerDependencies":
				insertErr = graph.InsertPeerDependency(neo4JDatabase, dependency)
			}
		}

		if insertErr != nil {
			log.Println("ERROR:", insertErr, "with job", j)
			// could access failure code from neo4j with .(messages.FailureMessage).Metadata["code"]
			jobs <- j
		}
		log.Println("worker", workerId, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", workerId)

	neo4JDatabase.Close()
}
