package plots

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"log"
	"os"
)

func CreateLinePlotForMaintainerPackageCount(maintainerName string, db *sql.DB) {
	maintainerCount, err := database.GetMaintainerCountsForMaintainer(maintainerName, db)
	if err != nil {
		log.Fatal(err)
	}
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = fmt.Sprintf("Package count evolution for maintainer: %v", maintainerName)
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{}
	p.Y.Label.Text = "Count"

	err = plotutil.AddLinePoints(p,
		maintainerName, GeneratePointsFromMaintainerCounts(maintainerCount))
	if err != nil {
		log.Fatal(err)
	}

	SavePlot(maintainerName, p)

	log.Printf("Finished maintainer %v", maintainerName)
}

type YearTicks struct{}

// Ticks computes the default tick marks, but inserts commas
// into the labels for the major tick marks.
func (YearTicks) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	val := 0.0
	for year := 2010; year < 2019; year++ {
		startMonth := 1
		endMonth := 12
		if year == 2010 {
			startMonth = 11
		}
		if year == 2018 {
			endMonth = 4
		}
		for month := startMonth; month <= endMonth; month++ {
			var tick plot.Tick
			if month == startMonth && year != 2010 {
				tick = plot.Tick{
					Value: val,
					Label: fmt.Sprintf("0%v.%v", startMonth, year),
				}
			} else {
				tick = plot.Tick{
					Value: val,
					Label: "",
				}
			}
			ticks = append(ticks, tick)
			val++
		}
	}
	return ticks
}

func SavePlot(maintainerName string, p *plot.Plot) {
	nestedDir := fmt.Sprintf("/home/markus/npm-analysis/maintainer-evolution/%v", string(maintainerName[0]))
	err := os.MkdirAll(nestedDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create nested directory with %v", err)
	}
	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("%v/maintainerPackageEvolution-%v.png", nestedDir, maintainerName)); err != nil {
		log.Fatal(err)
	}
}

func GeneratePointsFromMaintainerCounts(counts evolution.MaintainerCount) plotter.XYs {
	pts := make([]struct{ X, Y float64 }, 0)
	x := 0
	for year := 2010; year < 2019; year++ {
		startMonth := 1
		endMonth := 12
		if year == 2010 {
			startMonth = 11
		}
		if year == 2018 {
			endMonth = 4
		}
		for month := startMonth; month <= endMonth; month++ {
			y := counts.Counts[year][month]
			pts = append(pts, struct{ X, Y float64 }{X: float64(x), Y: float64(y)})
			x++
		}
	}
	return plotter.XYs(pts)
}
