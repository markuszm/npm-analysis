package plots

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func CreateLinePlotForMaintainerPackageCount(maintainerName string, db *sql.DB, createPlot bool) {
	maintainerCount, err := database.GetMaintainerCountsForMaintainer(maintainerName, db)
	if err != nil {
		log.Fatal(err)
	}
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = fmt.Sprintf("Package count evolution for maintainer: %v", maintainerName)
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{startYear: 2010}
	p.Y.Label.Text = "Count"

	points := GeneratePointsFromMaintainerCounts(maintainerCount)
	err = plotutil.AddLinePoints(p, points)
	if err != nil {
		log.Fatal(err)
	}

	if createPlot {
		SaveMaintainerPlot(maintainerName, "maintainer-evolution", p)
	} else {
		jsonBytes, err := json.Marshal(points)
		if err != nil {
			log.Fatal(err)
		}
		SaveValues(maintainerName, "maintainer-evolution", jsonBytes)
	}

	log.Printf("Finished maintainer %v", maintainerName)
}

func GenerateLinePlotForMaintainerReach(outputFolder string, maintainerName string, counts []float64, createPlot bool) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = fmt.Sprintf("Maintainer Reach Evolution for %v", maintainerName)
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{startYear: 2010}
	p.Y.Label.Text = "Reach"

	points := GeneratePointsFromFloatArray(counts)
	err = plotutil.AddLinePoints(p, points)
	if err != nil {
		log.Fatal(err)
	}

	if createPlot {
		SaveMaintainerPlot(maintainerName, outputFolder, p)
	} else {
		jsonBytes, err := json.Marshal(points)
		if err != nil {
			log.Fatal(err)
		}
		SaveValues(maintainerName, outputFolder, jsonBytes)
	}
}

func GenerateLinePlotForAverageResults(outputFolder, outputName string, counts []float64) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = fmt.Sprintf("Average Reach over all packages")
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{startYear: 2011}
	p.Y.Label.Text = "Reach"

	err = plotutil.AddLinePoints(p, GeneratePointsFromFloatArray(counts))
	if err != nil {
		log.Fatal(err)
	}

	SavePlot(path.Join(outputFolder, outputName+".png"), p)
}

func GenerateLinePlotForAverageMaintainerPackageCount(outputFolder string, counts []float64) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = fmt.Sprintf("Average Maintainer Package Count over all packages")
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{startYear: 2010}
	p.Y.Label.Text = "Package Count"

	err = plotutil.AddLinePoints(p, GeneratePointsFromFloatArray(counts))
	if err != nil {
		log.Fatal(err)
	}

	SavePlot(path.Join(outputFolder, "averageMaintainerPackageCount.png"), p)
}

func GenerateSortedLinePlotMaintainerPackageCount(outputFolder string, valuesPerYear map[int][]float64) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = fmt.Sprintf("All maintainers package count sorted by count")
	p.X.Label.Text = "Maintainer"
	p.Y.Label.Text = "Package Count"

	var args []interface{}
	for y := 2010; y < 2019; y++ {
		args = append(args, strconv.Itoa(y), GeneratePointsFromFloatArray(valuesPerYear[y]))
	}

	err = plotutil.AddLines(p, args...)
	if err != nil {
		log.Fatal(err)
	}

	SavePlot(path.Join(outputFolder, "sortedMaintainerPackageCount.png"), p)
}

type YearTicks struct {
	startYear int
}

func (y YearTicks) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	val := 0.0
	for year := y.startYear; year < 2019; year++ {
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

func SaveMaintainerPlot(maintainerName string, dir string, p *plot.Plot) {
	nestedDir := GetNestedDirName(maintainerName, dir)
	err := os.MkdirAll(nestedDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create nested directory with %v", err)
	}
	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 8*vg.Inch, GetPlotFileName(maintainerName, dir)); err != nil {
		log.Fatal(err)
	}
}

func SaveValues(name string, dir string, data []byte) {
	nestedDir := GetNestedDirName(name, dir)
	err := os.MkdirAll(nestedDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create nested directory with %v", err)
	}

	name = strings.Replace(name, "/", "", -1)
	name = strings.Replace(name, " ", "", -1)
	// Save the plot to a PNG file.
	filePath := fmt.Sprintf("%v/%v.json", nestedDir, name)

	err = ioutil.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func SavePlot(fileName string, p *plot.Plot) {
	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 8*vg.Inch, fileName); err != nil {
		log.Fatal(err)
	}
}

func GetNestedDirName(maintainerName string, dir string) string {
	return fmt.Sprintf("%v/%v", dir, string(maintainerName[0]))
}

func GetPlotFileName(maintainerName string, dir string) string {
	maintainerName = strings.Replace(maintainerName, "/", "", -1)
	maintainerName = strings.Replace(maintainerName, " ", "", -1)
	return fmt.Sprintf("%v/%v.png", GetNestedDirName(maintainerName, dir), maintainerName)
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

func GeneratePointsFromFloatArray(values []float64) plotter.XYs {
	pts := make([]struct{ X, Y float64 }, 0)
	for x, val := range values {
		pts = append(pts, struct{ X, Y float64 }{X: float64(x), Y: val})
	}
	return plotter.XYs(pts)
}
