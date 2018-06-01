package evolution

import (
	"github.com/markuszm/npm-analysis/model"
)

func CalculatePopularity(downloadCounts DownloadCountResponse) model.Popularity {

	overallDownloads := 0
	overallDays := 0
	year2015Downloads := 0
	year2015Days := 0
	year2016Downloads := 0
	year2016Days := 0
	year2017Downloads := 0
	year2017Days := 0
	year2018Downloads := 0
	year2018Days := 0

	for _, d := range downloadCounts.DownloadCounts {
		date := MustParseDate(d.Day)
		overallDownloads += d.Downloads
		overallDays++

		switch date.Year() {
		case 2015:
			year2015Downloads += d.Downloads
			year2015Days++

		case 2016:
			year2016Downloads += d.Downloads
			year2016Days++

		case 2017:
			year2017Downloads += d.Downloads
			year2017Days++

		case 2018:
			year2018Downloads += d.Downloads
			year2018Days++
		}
	}

	overall := float64(overallDownloads) / float64(overallDays)
	year2015 := float64(year2015Downloads) / float64(year2015Days)
	year2016 := float64(year2016Downloads) / float64(year2016Days)
	year2017 := float64(year2017Downloads) / float64(year2017Days)
	year2018 := float64(year2018Downloads) / float64(year2018Days)

	return model.Popularity{
		PackageName: downloadCounts.Package,
		Overall:     overall,
		Year2015:    year2015,
		Year2016:    year2016,
		Year2017:    year2017,
		Year2018:    year2018,
	}

}
