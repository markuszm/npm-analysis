package analysisimpl

type DeclaredDependenciesAnalysis struct {
}

func (d *DeclaredDependenciesAnalysis) AnalyzePackage(packagePath string) (string, error) {
	return "", nil
}

func (d *DeclaredDependenciesAnalysis) AnalyzePackages(packages map[string]string) (map[string]string, error) {
	results := make(map[string]string, len(packages))

	for pkg, pkgPath := range packages {
		result, err := d.AnalyzePackage(pkgPath)
		if err != nil {
			return results, err
		}
		results[pkg] = result
	}

	return results, nil
}
