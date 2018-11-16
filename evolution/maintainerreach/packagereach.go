package maintainerreach

func PackageReach(pkg string, dependentsMap map[string][]string, packages map[string]bool) {
	for _, dependent := range dependentsMap[pkg] {
		if ok := packages[dependent]; !ok {
			packages[dependent] = true
			PackageReach(dependent, dependentsMap, packages)
		}
	}
}

func PackageReachLayer(pkg string, dependentsMap map[string][]string, packages map[string]ReachDetails, layer int) {
	for _, dependent := range dependentsMap[pkg] {
		if _, ok := packages[dependent]; !ok {
			packages[dependent] = ReachDetails{layer, pkg}
			PackageReachLayer(dependent, dependentsMap, packages, layer+1)
		}
	}
}

type ReachDetails struct {
	Layer      int
	Dependency string
}
