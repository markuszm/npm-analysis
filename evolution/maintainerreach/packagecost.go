package maintainerreach

func PackageCost(pkg string, dependencyMap map[string]map[string]bool, packages map[string]bool) {
	for dependency, exists := range dependencyMap[pkg] {
		if !exists {
			continue
		}
		if ok := packages[dependency]; !ok {
			packages[dependency] = true
			PackageCost(dependency, dependencyMap, packages)
		}
	}
}
