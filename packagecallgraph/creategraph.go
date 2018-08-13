package packagecallgraph

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"log"
	"os"
	"sync"
)

type GraphCreator struct {
	neo4jUrl       string
	callgraphInput string
	exportsInput   string
	logger         *zap.SugaredLogger
	workers        int
}

func NewGraphCreator(neo4jUrl, callgraphInput, exportsInput string, workerNumber int, logger *zap.SugaredLogger) *GraphCreator {
	return &GraphCreator{neo4jUrl: neo4jUrl, callgraphInput: callgraphInput, exportsInput: exportsInput, workers: workerNumber, logger: logger}
}

func (g *GraphCreator) ExecCreation() error {
	err := g.initSchema()

	file, err := os.Open(g.callgraphInput)
	if err != nil {
		return errors.Wrap(err, "error opening callgraph.json file - does it exist in input folder?")
	}

	decoder := json.NewDecoder(file)

	workerWait := sync.WaitGroup{}

	jobs := make(chan model.PackageResult, 100)

	for w := 1; w <= g.workers; w++ {
		workerWait.Add(1)
		go g.worker(w, jobs, &workerWait)
	}

	for {
		result := model.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				log.Print("finished decoding result json")
				break
			} else {
				return errors.Wrap(err, "error processing package results")
			}
		}

		jobs <- result
	}

	return err
}

func (g *GraphCreator) initSchema() error {
	neo4JDatabase := graph.NewNeo4JDatabase()
	defer neo4JDatabase.Close()
	initErr := neo4JDatabase.InitDB(g.neo4jUrl)
	if initErr != nil {
		return initErr
	}
	graph.Init(neo4JDatabase)
	return nil
}

func (g *GraphCreator) worker(workerId int, jobs chan model.PackageResult, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	defer neo4JDatabase.Close()
	err := neo4JDatabase.InitDB(g.neo4jUrl)

	if err != nil {
		g.logger.Fatal(err)
	}

	for j := range jobs {
		calls, err := resultprocessing.TransformToCalls(j.Result)
		if err != nil {
			g.logger.Fatal(err)
		}
		g.logger.Debugf("Package: %s, Calls %v", j.Name, len(calls))
	}
	workerWait.Done()
}
