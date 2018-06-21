package analysisimpl

import (
	"bytes"
	"github.com/pkg/errors"
	"os/exec"
)

type AnalysisExecutor interface {
	AnalyzePackage(packagePath string) (string, error)
}

func ExecuteCommand(path string, args ...string) (string, error) {
	cmd := exec.Command(path, args...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, errOut.String())
	}
	return out.String(), nil
}

func AnalyzePackages(packages map[string]string, executor AnalysisExecutor) (map[string]string, error) {
	results := make(map[string]string, len(packages))

	for pkg, pkgPath := range packages {
		result, err := executor.AnalyzePackage(pkgPath)
		if err != nil {
			return results, err
		}
		results[pkg] = result
	}

	return results, nil
}
