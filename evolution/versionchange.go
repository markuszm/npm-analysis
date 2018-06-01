package evolution

import (
	"github.com/blang/semver"
	"sort"
	"time"
)

type VersionChange struct {
	PackageName string
	Version     string
	VersionPrev string
	VersionDiff string
	ReleaseTime time.Time
}

type VersionChanges []VersionChange

func (v VersionChanges) Len() int {
	return len(v)
}
func (v VersionChanges) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
func (v VersionChanges) Less(i, j int) bool {
	v1 := v[i]
	v2 := v[j]
	semver1 := semver.MustParse(v1.Version)
	semver2 := semver.MustParse(v2.Version)

	return semver1.LE(semver2)
}

func SortVersionChange(versionChanges []VersionChange) {
	changes := VersionChanges(versionChanges)
	sort.Sort(changes)
}
