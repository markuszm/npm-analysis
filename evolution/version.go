package evolution

import (
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"reflect"
	"time"
)

func ProcessVersions(metadata model.Metadata, timeCutoff time.Time) ([]VersionChange, error) {
	var changes []VersionChange

	versions := metadata.Versions
	var semvers semver.Versions
	for _, v := range versions {
		semverParsed := semver.MustParse(v.Version)
		semvers = append(semvers, semverParsed)
	}
	semver.Sort(semvers)

	lastVersion := ""
	var lastReleaseTime *time.Time = nil

	for _, s := range semvers {
		diff := ""
		if lastVersion == "" {
			diff = "publish"
		} else {
			diff = SemverDiff(semver.MustParse(lastVersion), s)
		}
		v := s.String()

		releaseTime := GetTimeForVersion(metadata, v)

		if releaseTime.After(timeCutoff) {
			continue
		}

		var timeDiff float64
		if lastReleaseTime == nil {
			timeDiff = 0.0
		} else {
			timeDiff = releaseTime.Sub(*lastReleaseTime).Hours()
		}

		change := VersionChange{
			PackageName: metadata.Name,
			Version:     v,
			VersionPrev: lastVersion,
			VersionDiff: diff,
			TimeDiff:    timeDiff,
			ReleaseTime: releaseTime,
		}
		changes = append(changes, change)
		lastVersion = v
		lastReleaseTime = &releaseTime
	}

	return changes, nil
}

func SemverDiff(a semver.Version, b semver.Version) string {
	if a.GT(b) {
		return "downgrade"
	}

	if !reflect.DeepEqual(a.Build, b.Build) {
		return "build"
	}

	if !reflect.DeepEqual(a.Pre, b.Pre) && len(b.Pre) > 0 {
		return "prerelease"
	}

	// also considers jump from prerelease to full release
	if a.Major != b.Major || (len(a.Pre) > 0 && len(b.Pre) == 0 && b.Minor == 0 && b.Patch == 0) {
		return "major"
	}

	if a.Minor != b.Minor || (len(a.Pre) > 0 && len(b.Pre) == 0 && b.Patch == 0) {
		return "minor"
	}

	if a.Patch != b.Patch || (len(a.Pre) > 0 && len(b.Pre) == 0) {
		return "patch"
	}

	return "equal"
}

func CountVersions(versionChanges []VersionChange) VersionCount {
	majorCount := 0
	minorBetweenMajorCount := 0
	patchBetweenMajorCount := 0
	patchBetweenMinorCount := 0
	minorTmp := 0
	minorCount := 0
	patchCount := 0
	patchMajorTmp := 0
	patchMinorTmp := 0
	for _, v := range versionChanges {
		switch v.VersionDiff {
		case "major":
			majorCount++
			minorBetweenMajorCount += minorTmp
			patchBetweenMajorCount += patchMajorTmp
			minorTmp = 0
			patchMajorTmp = 0
		case "minor":
			minorTmp++
			patchBetweenMinorCount += patchMinorTmp
			patchMinorTmp = 0
			minorCount++
		case "patch":
			patchMajorTmp++
			patchMinorTmp++
			patchCount++
		}
	}
	averageMinorsBetweenMajor := util.AvgInts(minorBetweenMajorCount, majorCount)
	averagePatchesBetweenMajor := util.AvgInts(patchBetweenMajorCount, majorCount)
	averagePatchesBetweenMinor := util.AvgInts(patchBetweenMinorCount, minorCount)
	versionCount := VersionCount{
		Major:                  majorCount,
		Minor:                  minorCount,
		Patch:                  patchCount,
		AvgMinorBetweenMajor:   averageMinorsBetweenMajor,
		AvgPatchesBetweenMajor: averagePatchesBetweenMajor,
		AvgPatchesBetweenMinor: averagePatchesBetweenMinor,
	}
	return versionCount
}

type VersionCount struct {
	Major, Minor, Patch    int
	AvgMinorBetweenMajor   float64
	AvgPatchesBetweenMajor float64
	AvgPatchesBetweenMinor float64
}
