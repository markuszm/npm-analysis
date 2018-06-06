package evolution

type MaintainerCount struct {
	Name   string
	Counts map[int]map[int]int
}

func CalculateMaintainerCounts(changes []MaintainerChange) map[string]MaintainerCount {
	maintainerCounts := make(map[string]MaintainerCount, 0)

	for _, c := range changes {
		m := maintainerCounts[c.Name]
		if m.Name == "" {
			maintainerCounts[c.Name] = MaintainerCount{
				Name:   c.Name,
				Counts: CreateCountMap(),
			}
		}

		year := c.ReleaseTime.Year()
		month := int(c.ReleaseTime.Month())
		if c.ChangeType == "INITIAL" || c.ChangeType == "ADDED" {
			for y := year; y < 2019; y++ {
				if y == year {
					for m := month; m <= 12; m++ {
						maintainerCounts[c.Name].Counts[y][m]++
					}
				} else {
					for m := 1; m <= 12; m++ {
						maintainerCounts[c.Name].Counts[y][m]++
					}
				}
			}
		}

		if c.ChangeType == "REMOVED" {
			for y := year; y < 2019; y++ {
				if y == year {
					for m := month + 1; m <= 12; m++ {
						maintainerCounts[c.Name].Counts[y][m]--
					}
				} else {
					for m := 1; m <= 12; m++ {
						maintainerCounts[c.Name].Counts[y][m]--
					}
				}
			}
		}
	}

	return maintainerCounts
}

func CreateCountMap() map[int]map[int]int {
	countMap := make(map[int]map[int]int, 9)

	for y := 2010; y <= 2018; y++ {
		countMap[y] = CreateMonthMap()
	}
	return countMap
}

func CreateMonthMap() map[int]int {
	monthMap := make(map[int]int, 12)
	for i := 1; i <= 12; i++ {
		monthMap[i] = 0
	}
	return monthMap
}
