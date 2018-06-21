package codeanalysis

import (
	"github.com/markuszm/npm-analysis/codeanalysis/analysisimpl"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"log"
	"os"
	"sync"
)

const deleteExtractedPackages = true

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

	log.Print("Successfully retrieved package names")

	results := make(map[string]string, len(packageNames))

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			log.Printf("Finished analyzing %d packages", i)
		}

		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			return "", err
		}
		results[pkg.Name] = result
	}

	log.Printf("Finished analyzing %v packages", len(packageNames))

	result, err = p.formatter.Format(results)
	return
}

func (p *Pipeline) ExecuteParallel(maxWorkers int) (result string, err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	log.Print("Successfully retrieved package names")

	workerGroup := sync.WaitGroup{}

	jobs := make(chan model.PackageVersionPair, 100)

	for i := 0; i < maxWorkers; i++ {
		workerGroup.Add(1)
		go p.worker(i, jobs, &workerGroup)
	}

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			log.Printf("Finished analyzing %d packages", i)
		}

		jobs <- pkg
	}

	close(jobs)

	workerGroup.Wait()

	log.Printf("Finished analyzing %v packages", len(packageNames))

	results := make(map[string]string, len(packageNames))
	p.resultMap.Range(func(key, value interface{}) bool {
		results[key.(string)] = value.(string)
		return true
	})
	result, err = p.formatter.Format(results)

	return
}

func (p *Pipeline) worker(workerId int, packages chan model.PackageVersionPair, workerGroup *sync.WaitGroup) {
	for pkg := range packages {
		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			// TODO: better error handling than panic
			log.Fatalf("FATAL ERROR with package %v: \n %v", pkg, err)
		}
		p.resultMap.Store(pkg.Name, result)
	}
	workerGroup.Done()
}

func (p *Pipeline) executePackageAnalysis(packageName model.PackageVersionPair) (result string, err error) {
	pkg, err := p.loader.LoadPackage(packageName)
	if err != nil {
		// if package is not found continue but result is an error message
		if err == ErrorNotFound {
			return ErrorNotFound.Error(), nil
		}
		err = errors.Wrap(err, "ERROR: loading package")
		return
	}

	packageFolderPath, err := p.unpacker.UnpackPackage(pkg)
	if err != nil {
		// TODO: more selective error management here after finding all possible errors
		result = err.Error()
		err = nil
		return
		//if !strings.Contains(err.Error(), "making hard link for") {
		//	err = errors.Wrap(err, "ERROR: unpacking package")
		//	return
		//}
	}

	result, err = p.analysis.AnalyzePackage(packageFolderPath)
	if err != nil {
		err = errors.Wrap(err, "ERROR: analyzing package")
		return
	}

	if deleteExtractedPackages {
		err = os.RemoveAll(packageFolderPath)
		if err != nil {
			err = errors.Wrap(err, "ERROR: removing tmp folder")
		}
	}
	return
}
