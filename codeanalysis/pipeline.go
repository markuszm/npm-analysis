package codeanalysis

import (
	"github.com/markuszm/npm-analysis/codeanalysis/analysisimpl"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"log"
	"os"
	"sync"
)

const cleanup = true

type Pipeline struct {
	collector NameCollector
	loader    PackageLoader
	unpacker  Unpacker
	analysis  analysisimpl.AnalysisExecutor
	writer    ResultWriter
}

func NewPipeline(collector NameCollector,
	loader PackageLoader,
	unpacker Unpacker,
	analysis analysisimpl.AnalysisExecutor,
	writer ResultWriter) *Pipeline {
	return &Pipeline{
		collector: collector,
		loader:    loader,
		unpacker:  unpacker,
		analysis:  analysis,
		writer:    writer,
	}
}

type PackageResult struct {
	Name    string
	Version string
	Result  interface{}
}

func (p *Pipeline) Execute() (err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	log.Print("Successfully retrieved package names")

	results := make(map[string]PackageResult, len(packageNames))

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			log.Printf("Finished analyzing %d packages", i)
		}

		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			return err
		}
		pkgResult := PackageResult{Name: pkg.Name, Version: pkg.Version, Result: result}
		results[pkg.Name] = pkgResult
	}

	log.Printf("Finished analyzing %v packages", len(packageNames))

	err = p.writer.WriteAll(results)
	return
}

func (p *Pipeline) ExecuteParallel(maxWorkers int) (err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	log.Print("Successfully retrieved package names")

	jobGroup := sync.WaitGroup{}
	resultGroup := sync.WaitGroup{}

	jobs := make(chan model.PackageVersionPair, 100)
	resultsChan := make(chan PackageResult, 1000)

	resultGroup.Add(1)
	go p.writer.WriteBuffered(resultsChan, &resultGroup)

	for i := 0; i < maxWorkers; i++ {
		jobGroup.Add(1)
		go p.worker(i, jobs, resultsChan, &jobGroup)
	}

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			log.Printf("Finished analyzing %d packages", i)
		}
		jobs <- pkg
	}

	close(jobs)
	jobGroup.Wait()

	close(resultsChan)
	resultGroup.Wait()

	log.Printf("Finished analyzing %v packages", len(packageNames))
	return
}

func (p *Pipeline) worker(workerId int, packages chan model.PackageVersionPair, results chan PackageResult, workerGroup *sync.WaitGroup) {
	for pkg := range packages {
		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			// TODO: better error handling than panic
			log.Fatalf("FATAL ERROR with package %v: \n %v", pkg, err)
		}
		pkgResult := PackageResult{Name: pkg.Name, Version: pkg.Version, Result: result}
		results <- pkgResult
	}
	workerGroup.Done()
}

func (p *Pipeline) executePackageAnalysis(packageName model.PackageVersionPair) (result interface{}, err error) {
	pkg, err := p.loader.LoadPackage(packageName)
	if err != nil {
		// if package is not found continue but result is an error message
		if err == ErrorNotFound {
			return []string{ErrorNotFound.Error()}, nil
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

	if cleanup {
		err = os.RemoveAll(packageFolderPath)
		if err != nil {
			err = errors.Wrap(err, "ERROR: removing tmp folder")
		}

		if p.loader.NeedsCleanup() {
			err := os.Remove(pkg)

			if err != nil {
				err = errors.Wrap(err, "ERROR: removing tmp package file")
			}
		}
	}
	return
}
