package evolution

import (
	"github.com/markuszm/npm-analysis/model"
	"time"
)

func generatePackageData(metadata model.Metadata, date time.Time) PackageData {
	latestVersion := FindLatestVersion(metadata, date)
	if latestVersion == "unreleased" {
		packageData := PackageData{
			Version:      latestVersion,
			Dependencies: []string{},
			Maintainers:  []string{},
		}
		return packageData
	}
	versionMetadata := metadata.Versions[latestVersion]
	maintainers := ParseMaintainers(versionMetadata.Maintainers)
	dependencies := versionMetadata.Dependencies
	var dependencyNames []string
	for d, _ := range dependencies {
		dependencyNames = append(dependencyNames, d)
	}
	var maintainerNames []string
	for _, m := range maintainers {
		maintainerNames = append(maintainerNames, m.Name)
	}
	packageData := PackageData{
		Version:      latestVersion,
		Dependencies: dependencyNames,
		Maintainers:  maintainerNames,
	}
	return packageData
}

type PackageData struct {
	Version      string   `json:"version"`
	Dependencies []string `json:"dependencies"`
	Maintainers  []string `json:"maintainers"`
}

// start time is 11.2010 and cutoff is 04.2018
func GetPackageMetadataForEachMonth(metadata model.Metadata) map[time.Time]PackageData {
	resultMap := make(map[time.Time]PackageData, 0)
	for y := 2010; y <= 2018; y++ {
		startMonth := 1
		endMonth := 12
		if y == 2010 {
			startMonth = 11
		}

		if y == 2018 {
			endMonth = 4
		}

		for m := startMonth; m <= endMonth; m++ {
			date := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
			packageData := generatePackageData(metadata, date)
			resultMap[date] = packageData
		}
	}
	return resultMap
}
