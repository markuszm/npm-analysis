package util

import "time"

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type TimeValue struct {
	Key   time.Time
	Value float64
}

type TimeValueList []TimeValue

func (p TimeValueList) Len() int           { return len(p) }
func (p TimeValueList) Less(i, j int) bool { return p[i].Key.Before(p[j].Key) }
func (p TimeValueList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type FloatList []float64

func (p FloatList) Len() int           { return len(p) }
func (p FloatList) Less(i, j int) bool { return p[i] < p[j] }
func (p FloatList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type MaintainerReachDiff struct {
	Name string
	Diff float64
}

type MaintainerReachDiffList []MaintainerReachDiff

func (m MaintainerReachDiffList) Len() int           { return len(m) }
func (m MaintainerReachDiffList) Less(i, j int) bool { return m[i].Diff < m[j].Diff }
func (m MaintainerReachDiffList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }