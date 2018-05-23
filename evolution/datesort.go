package evolution

import (
	"sort"
	"time"
)

type VersionTimePair struct {
	Key   string
	Value time.Time
}

// A slice of pairs that implements sort.Interface to sort by values
type VersionTimeList []VersionTimePair

func (p VersionTimeList) Len() int           { return len(p) }
func (p VersionTimeList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p VersionTimeList) Less(i, j int) bool { return p[i].Value.Before(p[j].Value) }

func SortTime(times map[string]time.Time) VersionTimeList {
	var r []VersionTimePair
	for k, v := range times {
		r = append(r, VersionTimePair{Key: k, Value: v})
	}
	sort.Sort(VersionTimeList(r))
	return r
}
