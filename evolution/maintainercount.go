package evolution

type MaintainerCount struct {
	Name   string
	Counts map[int]int
}

func CalculateMaintainerCountPerYear(changes []MaintainerChange) map[string]MaintainerCount {
	maintainerCounts := make(map[string]MaintainerCount, 0)

	for _, c := range changes {
		m := maintainerCounts[c.Name]
		if m.Name == "" {
			maintainerCounts[c.Name] = MaintainerCount{
				Name: c.Name,
				Counts: map[int]int{
					2010: 0,
					2011: 0,
					2012: 0,
					2013: 0,
					2014: 0,
					2015: 0,
					2016: 0,
					2017: 0,
					2018: 0,
				},
			}
		}

		year := c.ReleaseTime.Year()
		if c.ChangeType == "INITIAL" || c.ChangeType == "ADDED" {
			for i := year; i < 2019; i++ {
				maintainerCounts[c.Name].Counts[i]++
			}
		}

		if c.ChangeType == "REMOVED" {
			for i := year + 1; i < 2019; i++ {
				maintainerCounts[c.Name].Counts[i]--
			}
		}
	}

	return maintainerCounts
}
