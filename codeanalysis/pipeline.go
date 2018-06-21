package codeanalysis

import (
	"github.com/markuszm/npm-analysis/codeanalysis/analysisimpl"
	"github.com/pkg/errors"
	"log"
	"sync"
)

type Pipeline struct {
	collector NameCollector
	loader    PackageLoader
	unpacker  Unpacker
	analysis  analysisimpl.AnalysisExecutor
	formatter ResultFormatter
	resultMap sync.Map
}

func NewPipeline(collector NameCollector,
	loader PackageLoader,
	unpacker Unpacker,
	analysis analysisimpl.AnalysisExecutor,
	formatter ResultFormatter) *Pipeline {
	return &Pipeline{
		collector: collector,
		loader:    loader,
		unpacker:  unpacker,
		analysis:  analysis,
		formatter: formatter,
		resultMap: sync.Map{},
	}
}

func (p *Pipeline) Execute() (result string, err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	packages, err := p.loader.LoadPackages(packageNames)
	if err != nil {
		err = errors.Wrap(err, "ERROR: loading packages")
		return
	}

	packages, err = p.unpacker.UnpackPackages(packages)
	if err != nil {
		err = errors.Wrap(err, "ERROR: unpacking packages")
		return
	}

	results, err := p.analysis.AnalyzePackages(packages)
	if err != nil {
		err = errors.Wrap(err, "ERROR: analyzing packages")
		return
	}

	result, err = p.formatter.Format(results)
	return
}

func (p *Pipeline) ExecuteParallel(maxWorkers int) (result string, err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for i := 0; i < maxWorkers; i++ {
		go p.worker(i, jobs, &workerWait)
	}

	for _, pkg := range packageNames {
		jobs <- pkg
	}

	close(jobs)

	workerWait.Wait()

	results := make(map[string]string, len(packageNames))
	p.resultMap.Range(func(key, value interface{}) bool {
		results[key.(string)] = value.(string)
		return true
	})
	result, err = p.formatter.Format(results)

	return
}

func (p *Pipeline) worker(workerId int, packages chan string, group *sync.WaitGroup) {
	for pkg := range packages {
		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			// TODO: better error handling then just fatal
			log.Fatalf("FATAL ERROR with package %v: \n %v", pkg, err)
		}
		p.resultMap.Store(pkg, result)
	}
	group.Done()
}

func (p *Pipeline) executePackageAnalysis(packageName string) (result string, err error) {
	packages, err := p.loader.LoadPackage(packageName)
	if err != nil {
		err = errors.Wrap(err, "ERROR: loading packages")
		return
	}

	packages, err = p.unpacker.UnpackPackage(packages)
	if err != nil {
		err = errors.Wrap(err, "ERROR: unpacking packages")
		return
	}

	result, err = p.analysis.AnalyzePackage(packages)
	if err != nil {
		err = errors.Wrap(err, "ERROR: analyzing packages")
		return
	}
	return
}
