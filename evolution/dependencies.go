package evolution

import (
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/model"
	"time"
)

type DependencyChange struct {
	PackageName           string
	PackageVersion        string
	DependencyName        string
	DependencyVersion     string
	DependencyVersionPrev string
	ReleaseTime           time.Time
	ChangeType            string
}

func ProcessDependencies(metadata model.Metadata, timeCutoff time.Time) ([]DependencyChange, error) {
	var changeList []DependencyChange
	dependenciesSet := make(map[string]string)
	versions := metadata.Versions
	var semvers semver.Versions
	for _, v := range versions {
		semverParsed := semver.MustParse(v.Version)
		semvers = append(semvers, semverParsed)
	}
	semver.Sort(semvers)

	for i, s := range semvers {
		pkgVer := s.String()
		pkgData := versions[pkgVer]
		dependencies := pkgData.Dependencies
		seenDependencies := make(map[string]bool, 0)
		releaseTime := GetTimeForVersion(metadata, pkgVer)

		if releaseTime.After(timeCutoff) {
			continue
		}

		for d, v := range dependencies {
			parsedVer := ParseVersion(v)

			// dependency added
			if dependenciesSet[d] == "" {
				changeType := "ADDED"
				// if first existing version then changeType is initial instead
				if i == 0 {
					changeType = "INITIAL"
				}
				dependencyChange := DependencyChange{
					PackageName:           metadata.Name,
					PackageVersion:        pkgVer,
					DependencyName:        d,
					DependencyVersion:     parsedVer,
					DependencyVersionPrev: "",
					ReleaseTime:           releaseTime,
					ChangeType:            changeType,
				}
				changeList = append(changeList, dependencyChange)
				seenDependencies[d] = true
				dependenciesSet[d] = parsedVer
			} else {
				oldVersion := dependenciesSet[d]
				// dependency updated when versions differ
				if oldVersion != parsedVer {
					dependencyChange := DependencyChange{
						PackageName:           metadata.Name,
						PackageVersion:        pkgVer,
						DependencyName:        d,
						DependencyVersion:     parsedVer,
						DependencyVersionPrev: oldVersion,
						ReleaseTime:           releaseTime,
						ChangeType:            "UPDATED",
					}
					changeList = append(changeList, dependencyChange)
					dependenciesSet[d] = parsedVer
				}
				seenDependencies[d] = true
			}
		}

		for d, v := range dependenciesSet {
			if v == "" {
				continue
			}
			// dependency removed
			if !seenDependencies[d] {
				dependencyChange := DependencyChange{
					PackageName:           metadata.Name,
					PackageVersion:        pkgVer,
					DependencyName:        d,
					DependencyVersion:     v,
					DependencyVersionPrev: v,
					ReleaseTime:           releaseTime,
					ChangeType:            "REMOVED",
				}
				changeList = append(changeList, dependencyChange)
				dependenciesSet[d] = ""
			}
		}
	}
	return changeList, nil
}

func ParseVersion(version interface{}) (parsedVersion string) {
	if version == nil {
		return
	}
	switch version.(type) {
	case string:
		parsedVersion = version.(string)
	case map[string]interface{}:
		depMap := version.(map[string]interface{})
		nestedVersion := depMap["version"]
		if nestedVersion != nil {
			parsedVersion = nestedVersion.(string)
		}
	}
	return
}
