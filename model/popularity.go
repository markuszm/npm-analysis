package model

import (
	"github.com/markuszm/npm-analysis/util"
)

type Popularity struct {
	PackageName                                     string
	Overall, Year2015, Year2016, Year2017, Year2018 float64
}

type PopularityMonthly struct {
	PackageName string
	Popularity  []util.TimePopularity
}
