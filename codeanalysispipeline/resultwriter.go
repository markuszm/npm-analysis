package codeanalysispipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type ResultWriter interface {
	WriteAll(results map[string]model.PackageResult) error
	WriteBuffered(results chan model.PackageResult, workerGroup *sync.WaitGroup) error
}

type CSVWriter struct {
	FilePath string
}

func NewCSVWriter(filePath string) *CSVWriter {
	return &CSVWriter{FilePath: filePath}
}

func (c *CSVWriter) WriteAll(results map[string]model.PackageResult) error {
	file, err := os.Create(c.FilePath)

	if err != nil {
		log.Fatal("Cannot create result file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, r := range results {
		var fields []string
		fields = append(fields, r.Name, r.Version, fmt.Sprint(r.Result))
		err = writer.Write(fields)
		if err != nil {
			log.Fatal("Cannot write to result file", err)
		}
	}

	return nil
}

func (c *CSVWriter) WriteBuffered(results chan model.PackageResult, workerGroup *sync.WaitGroup) error {
	file, err := os.Create(c.FilePath)

	if err != nil {
		log.Fatal("Cannot create result file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	defer writer.Flush()
	i := 0

	for r := range results {
		var fields []string
		fields = append(fields, r.Name, r.Version, fmt.Sprint(r.Result))
		err = writer.Write(fields)
		if err != nil {
			log.Fatal("Cannot write to result file", err)
		}
		if i%1000 == 0 {
			writer.Flush()
		}
		i++
	}

	workerGroup.Done()

	return nil
}

type JSONWriter struct {
	FilePath string
}

func NewJSONWriter(filePath string) *JSONWriter {
	return &JSONWriter{FilePath: filePath}
}

func (j *JSONWriter) WriteAll(results map[string]model.PackageResult) error {
	// TODO: needs to also use json decoder else all result processing fails
	bytes, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return errors.Wrap(err, "error marshalling results as json")
	}
	err = ioutil.WriteFile(j.FilePath, bytes, os.ModePerm)
	return err
}

func (j *JSONWriter) WriteBuffered(results chan model.PackageResult, workerGroup *sync.WaitGroup) error {
	file, err := os.Create(j.FilePath)

	if err != nil {
		log.Fatal("Cannot create result file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	for r := range results {
		err := encoder.Encode(r)
		if err != nil {
			log.Fatal("Cannot write to result file", err)
		}

	}

	workerGroup.Done()

	return nil
}
