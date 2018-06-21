package analysisimpl

type DeclaredDependenciesAnalysis struct {
}

func (d *DeclaredDependenciesAnalysis) AnalyzePackage(packagePath string) (string, error) {
	return "", nil
}

func (d *DeclaredDependenciesAnalysis) AnalyzePackages(packages map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(packages))
	return result, nil
}
