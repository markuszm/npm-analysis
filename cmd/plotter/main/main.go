package main

import (
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"log"
	"math/rand"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if err != nil {
		log.Fatal(err)
	}
	defer mysql.Close()

	maintainerName := "types"
	maintainerCount, err := database.GetMaintainerCountsForMaintainer(maintainerName, mysql)
	if err != nil {
		log.Fatal(err)
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = fmt.Sprintf("Package count evolution for maintainer: %v", maintainerName)
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Count"

	err = plotutil.AddLinePoints(p,
		maintainerName, generatePointsFromMaintainerCounts(maintainerCount))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "maintainerPackageEvolution.png"); err != nil {
		panic(err)
	}

}

func generatePointsFromMaintainerCounts(counts evolution.MaintainerCount) plotter.XYs {
	pts := make([]struct{ X, Y float64 }, 0)
	x := 0
	for _, innerMap := range counts.Counts {
		for _, count := range innerMap {
			pts = append(pts, struct{ X, Y float64 }{X: float64(x), Y: float64(count)})
			x++
		}
	}
	return plotter.XYs(pts)
}

func boxPlotExample() {
	// Get some data to display in our plot.
	rand.Seed(int64(0))
	n := 10
	uniform := make(plotter.Values, n)
	normal := make(plotter.Values, n)
	expon := make(plotter.Values, n)
	for i := 0; i < n; i++ {
		uniform[i] = rand.Float64()
		normal[i] = rand.NormFloat64()
		expon[i] = rand.ExpFloat64()
	}
	// Create the plot and set its title and axis label.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Box plots"
	p.Y.Label.Text = "Values"
	// Make boxes for our data and add them to the plot.
	w := vg.Points(20)
	b0, err := plotter.NewBoxPlot(w, 0, uniform)
	if err != nil {
		panic(err)
	}
	b1, err := plotter.NewBoxPlot(w, 1, normal)
	if err != nil {
		panic(err)
	}
	b2, err := plotter.NewBoxPlot(w, 2, expon)
	if err != nil {
		panic(err)
	}
	p.Add(b0, b1, b2)
	// Set the X axis of the plot to nominal with
	// the given names for x=0, x=1 and x=2.
	p.NominalX("Uniform\nDistribution", "Normal\nDistribution",
		"Exponential\nDistribution")
	if err := p.Save(3*vg.Inch, 4*vg.Inch, "boxplot.png"); err != nil {
		panic(err)
	}
}
