package codeanalysis

import (
	"github.com/pkg/errors"
	"log"
)

type Pipeline struct {
	collector NameCollector
}

func NewPipeline(collector NameCollector) *Pipeline {
	return &Pipeline{collector: collector}
}

func (p *Pipeline) Execute() (string, error) {
	var result string

	packageNames, err := p.collector.GetPackageNames()
	if err != nil {
		return result, errors.Wrap(err, "ERROR: retrieving package names")
	}
	for _, p := range packageNames {
		log.Print(p)
	}
	return result, nil
}
