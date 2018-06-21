package analysisimpl

type AnalysisExecutor interface {
	AnalyzePackage(packagePath string) (string, error)
	AnalyzePackages(packages map[string]string) (map[string]string, error)
}
