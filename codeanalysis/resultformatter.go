package codeanalysis

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type ResultFormatter interface {
	Format(results map[string]string) (string, error)
}

type CSVFormatter struct {
}

func (c *CSVFormatter) Format(results map[string]string) (string, error) {
	// TODO: if needed implement csv formatter
	return "", nil
}

type JSONFormatter struct {
}

func (j *JSONFormatter) Format(results map[string]string) (string, error) {
	bytes, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return "", errors.Wrap(err, "error marshalling results as json")
	}
	return string(bytes), nil
}
