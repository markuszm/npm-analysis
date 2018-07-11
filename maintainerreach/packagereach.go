package maintainerreach

func PackageReach(pkg string, dependentsMap map[string][]string, packages map[string]bool) {
	for _, dependent := range dependentsMap[pkg] {
		if ok := packages[dependent]; !ok {
			packages[dependent] = true
			PackageReach(dependent, dependentsMap, packages)
		}
	}
}

func PackageReachLayer(pkg string, dependentsMap map[string][]string, packages map[string]int, layer int) {
	for _, dependent := range dependentsMap[pkg] {
		if ok := packages[dependent]; ok == 0 {
			packages[dependent] = layer
			PackageReachLayer(dependent, dependentsMap, packages, layer+1)
		}
	}
}
