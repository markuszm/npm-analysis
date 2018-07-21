package codeanalysis

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"os/exec"
	"time"
)

type AnalysisExecutor interface {
	AnalyzePackage(packagePath string) (interface{}, error)
}

func ExecuteCommand(path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("command timed out")
	}

	if err != nil {
		return "", errors.Wrapf(err, errOut.String())
	}
	return out.String(), nil
}

func AnalyzePackages(packages map[string]string, executor AnalysisExecutor) (map[string]interface{}, error) {
	results := make(map[string]interface{}, len(packages))

	for pkg, pkgPath := range packages {
		result, err := executor.AnalyzePackage(pkgPath)
		if err != nil {
			return results, err
		}
		results[pkg] = result
	}

	return results, nil
}
