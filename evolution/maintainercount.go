package evolution

type MaintainerCount struct {
	Name   string
	Counts map[int]map[int]int
}

type MaintainedPackages struct {
	Name     string
	Packages map[int]map[int]PackageSet
}

// start time is 11.2010 and cutoff is 04.2018
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
			for y := year; y <= 2018; y++ {
				startMonth := 1
				endMonth := 12
				if y == 2010 {
					startMonth = 11
				}

				if y == year {
					startMonth = month
				}

				if y == 2018 {
					endMonth = 4
				}

				for m := startMonth; m <= endMonth; m++ {
					maintainerCounts[c.Name].Counts[y][m]++

				}
			}
		}

		if c.ChangeType == "REMOVED" {
			for y := year; y <= 2018; y++ {
				startMonth := 1
				endMonth := 12
				if y == 2010 {
					startMonth = 11
				}

				if y == year {
					startMonth = month
				}

				if y == 2018 {
					endMonth = 4
				}

				for m := startMonth; m <= endMonth; m++ {
					maintainerCounts[c.Name].Counts[y][m]--

				}
			}
		}
	}

	return maintainerCounts
}

func CalculateMaintainerPackages(changes []MaintainerChange) map[string]MaintainedPackages {
	maintainers := make(map[string]MaintainedPackages, 0)

	for _, c := range changes {
		m := maintainers[c.Name]
		if m.Name == "" {
			maintainers[c.Name] = MaintainedPackages{
				Name:     c.Name,
				Packages: CreatePackageMap(),
			}
		}

		year := c.ReleaseTime.Year()
		month := int(c.ReleaseTime.Month())
		if c.ChangeType == "INITIAL" || c.ChangeType == "ADDED" {
			for y := year; y <= 2018; y++ {
				startMonth := 1
				endMonth := 12
				if y == 2010 {
					startMonth = 11
				}

				if y == year {
					startMonth = month
				}

				if y == 2018 {
					endMonth = 4
				}

				for m := startMonth; m <= endMonth; m++ {
					maintainers[c.Name].Packages[y][m][c.PackageName] = true

				}
			}
		}

		if c.ChangeType == "REMOVED" {
			for y := year; y <= 2018; y++ {
				startMonth := 1
				endMonth := 12
				if y == 2010 {
					startMonth = 11
				}

				if y == year {
					startMonth = month
				}

				if y == 2018 {
					endMonth = 4
				}

				for m := startMonth; m <= endMonth; m++ {
					maintainers[c.Name].Packages[y][m][c.PackageName] = false

				}
			}
		}
	}

	return maintainers
}

type PackageSet map[string]bool

func CreatePackageMap() map[int]map[int]PackageSet {
	countMap := make(map[int]map[int]PackageSet, 9)

	for y := 2010; y <= 2018; y++ {
		countMap[y] = CreateMonthPackageMap(y)
	}
	return countMap
}

func CreateCountMap() map[int]map[int]int {
	countMap := make(map[int]map[int]int, 9)

	for y := 2010; y <= 2018; y++ {
		countMap[y] = CreateMonthCountMap(y)
	}
	return countMap
}

func CreateMonthCountMap(year int) map[int]int {
	monthMap := make(map[int]int, 0)
	startMonth := 1
	endMonth := 12
	if year == 2010 {
		startMonth = 11
	}
	if year == 2018 {
		endMonth = 4
	}

	for m := startMonth; m <= endMonth; m++ {
		monthMap[m] = 0
	}
	return monthMap
}

func CreateMonthPackageMap(year int) map[int]PackageSet {
	monthMap := make(map[int]PackageSet, 0)
	startMonth := 1
	endMonth := 12
	if year == 2010 {
		startMonth = 11
	}
	if year == 2018 {
		endMonth = 4
	}

	for m := startMonth; m <= endMonth; m++ {
		monthMap[m] = make(PackageSet, 0)
	}
	return monthMap
}
