package codeanalysis

type ResultFormatter interface {
	Format(results map[string]string) (string, error)
}

type CSVFormatter struct {
}

func (c *CSVFormatter) Format(results map[string]string) (string, error) {
	// TODO
	return "", nil
}

type JSONFormatter struct {
}

func (j *JSONFormatter) Format(results map[string]string) (string, error) {
	// TODO
	return "", nil
}
