package codeanalysispipeline

import (
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"sync"
)

type Pipeline struct {
	collector NameCollector
	loader    PackageLoader
	unpacker  Unpacker
	analysis  codeanalysis.AnalysisExecutor
	writer    ResultWriter
	logger    *zap.SugaredLogger
	cleanup   bool
}

func NewPipeline(collector NameCollector,
	loader PackageLoader,
	unpacker Unpacker,
	analysis codeanalysis.AnalysisExecutor,
	writer ResultWriter,
	logger *zap.SugaredLogger,
	cleanup bool) *Pipeline {
	return &Pipeline{
		collector: collector,
		loader:    loader,
		unpacker:  unpacker,
		analysis:  analysis,
		writer:    writer,
		logger:    logger,
		cleanup:   cleanup,
	}
}

func (p *Pipeline) Execute() (err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	p.logger.Info("Successfully retrieved package names")

	results := make(map[string]model.PackageResult, len(packageNames))

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			p.logger.Infof("Finished analyzing %d packages", i)
		}

		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			return err
		}
		pkgResult := model.PackageResult{Name: pkg.Name, Version: pkg.Version, Result: result}
		results[pkg.Name] = pkgResult
	}

	p.logger.Infof("Finished analyzing %v packages", len(packageNames))

	err = p.writer.WriteAll(results)
	return
}

func (p *Pipeline) ExecuteParallel(maxWorkers int) (err error) {
	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		err = errors.Wrap(err, "ERROR: retrieving package names")
		return
	}

	p.logger.Infof("Successfully retrieved package names")

	jobGroup := sync.WaitGroup{}
	resultGroup := sync.WaitGroup{}

	jobs := make(chan model.PackageVersionPair, 100)
	resultsChan := make(chan model.PackageResult, 1000)

	resultGroup.Add(1)
	go p.writer.WriteBuffered(resultsChan, &resultGroup)

	for i := 0; i < maxWorkers; i++ {
		jobGroup.Add(1)
		go p.worker(i, jobs, resultsChan, &jobGroup)
	}

	for i, pkg := range packageNames {
		if i%1000 == 0 {
			p.logger.Infof("Finished analyzing %d packages", i)
		}
		jobs <- pkg
	}

	close(jobs)
	jobGroup.Wait()

	close(resultsChan)
	resultGroup.Wait()

	p.logger.Infof("Finished analyzing %v packages", len(packageNames))
	return
}

func (p *Pipeline) worker(workerId int, packages chan model.PackageVersionPair, results chan model.PackageResult, workerGroup *sync.WaitGroup) {
	for pkg := range packages {
		result, err := p.executePackageAnalysis(pkg)
		if err != nil {
			errorStr := fmt.Sprintf("ERROR with package %v: \n %v", pkg, err)
			p.logger.Errorf(errorStr)
			pkgResult := model.PackageResult{Name: pkg.Name, Version: pkg.Version, Result: errorStr}
			results <- pkgResult
			continue
		}
		pkgResult := model.PackageResult{Name: pkg.Name, Version: pkg.Version, Result: result}
		results <- pkgResult
	}
	workerGroup.Done()
}

func (p *Pipeline) executePackageAnalysis(packageName model.PackageVersionPair) (result interface{}, err error) {
	pkg, err := p.loader.LoadPackage(packageName)
	if err != nil {
		err = errors.Wrap(err, "loading package")
		return
	}

	packageFolderPath, err := p.unpacker.UnpackPackage(pkg)
	if err != nil {
		p.logger.Errorf("unpacking package %v with %v", packageName, err)
		err = errors.Wrap(err, "unpacking package")
		return
		//if !strings.Contains(err.Error(), "making hard link for") {
		//	err = errors.Wrap(err, "ERROR: unpacking package")
		//	return
		//}
	}

	result, err = p.analysis.AnalyzePackage(packageFolderPath)
	if err != nil {
		err = errors.Wrap(err, "analyzing package")
		return
	}

	if p.cleanup {
		err = os.RemoveAll(packageFolderPath)
		if err != nil {
			err = errors.Wrap(err, "removing tmp folder")
		}

		if p.loader.NeedsCleanup() {
			err := os.Remove(pkg)

			if err != nil {
				err = errors.Wrap(err, "removing tmp package file")
			}
		}
	}
	return
}
