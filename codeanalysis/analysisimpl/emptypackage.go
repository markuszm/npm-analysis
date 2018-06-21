package analysisimpl

type EmptyPackageAnalysis struct {
}

func (e *EmptyPackageAnalysis) AnalyzePackage(packagePath string) (string, error) {
	return "", nil
}

func (e *EmptyPackageAnalysis) AnalyzePackages(packages map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(packages))
	return result, nil
}
