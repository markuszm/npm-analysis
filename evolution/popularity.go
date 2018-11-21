package evolution

import (
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"sort"
	"time"
)

func CalculatePopularityByYear(pkgName string, downloadCounts DownloadCountResponse) model.Popularity {

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

	overall := util.AvgInts(overallDownloads, overallDays)
	year2015 := util.AvgInts(year2015Downloads, year2015Days)
	year2016 := util.AvgInts(year2016Downloads, year2016Days)
	year2017 := util.AvgInts(year2017Downloads, year2017Days)
	year2018 := util.AvgInts(year2018Downloads, year2018Days)

	return model.Popularity{
		PackageName: pkgName,
		Overall:     overall,
		Year2015:    year2015,
		Year2016:    year2016,
		Year2017:    year2017,
		Year2018:    year2018,
	}

}

func CalculatePopularityByMonth(pkgName string, downloadCounts DownloadCountResponse) model.PopularityMonthly {
	timeDownloadsMap := make(map[time.Time]float64, 0)

	// this assumes that first month of download count response is January 2015
	currentMonth := time.Month(1)
	currentYear := 2015
	currentDays := 0
	currentDownloadSum := 0

	for _, d := range downloadCounts.DownloadCounts {
		downloadDate := MustParseDate(d.Day)
		month := downloadDate.Month()
		if currentMonth != month {
			date := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)

			timeDownloadsMap[date] = util.AvgInts(currentDownloadSum, currentDays)

			currentMonth = month
			currentYear = downloadDate.Year()
			currentDays = 1
			currentDownloadSum = d.Downloads
		} else {
			currentDays++
			currentDownloadSum += d.Downloads
		}
	}

	// add remaining month
	date := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
	timeDownloadsMap[date] = util.AvgInts(currentDownloadSum, currentDays)

	var timePopularityList []util.TimePopularity
	// sort time map
	for t, d := range timeDownloadsMap {
		timePopularityList = append(timePopularityList, util.TimePopularity{Time: t, Downloads: d})
	}
	sort.Sort(util.TimePopularityList(timePopularityList))

	return model.PopularityMonthly{
		PackageName: pkgName,
		Popularity:  timePopularityList,
	}
}
