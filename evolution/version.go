package evolution

import (
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/model"
	"reflect"
	"time"
)

type VersionChange struct {
	PackageName string
	Version     string
	VersionPrev string
	VersionDiff string
	ReleaseTime time.Time
}

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

		change := VersionChange{
			PackageName: metadata.Name,
			Version:     v,
			VersionPrev: lastVersion,
			VersionDiff: diff,
			ReleaseTime: releaseTime,
		}
		changes = append(changes, change)
		lastVersion = v
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

	// edge case jump from prerelease to next major
	if a.Major != b.Major || (len(a.Pre) > 1 && len(b.Pre) == 0) {
		return "major"
	}

	if a.Minor != b.Minor {
		return "minor"
	}

	if a.Patch != b.Patch {
		return "patch"
	}

	return "equal"
}
