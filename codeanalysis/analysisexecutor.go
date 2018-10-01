package codeanalysis

import (
	"bytes"
	"context"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"os/exec"
	"time"
)

type AnalysisExecutor interface {
	AnalyzePackageFiles(packagePath string) (interface{}, error)
	AnalyzePackage(version model.PackageVersionPair) (interface{}, error)
}

func ExecuteCommand(path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
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
		result, err := executor.AnalyzePackageFiles(pkgPath)
		if err != nil {
			return results, err
		}
		results[pkg] = result
	}

	return results, nil
}
