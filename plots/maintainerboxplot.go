package plots

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"log"
	"math"
	"time"
)

func CreateBoxPlot(values map[time.Time][]int) {
	var allValues []plotter.Values
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
			counts := values[time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)]
			var plotterValues []float64
			for _, c := range counts {
				val := math.Log10(float64(c) + 0.01)
				plotterValues = append(plotterValues, val)
			}
			allValues = append(allValues, plotter.Values(plotterValues))
		}
	}

	// Create the plot and set its title and axis label.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Box plots"
	p.Y.Label.Text = "Package Count"
	// Make boxes for our data and add them to the plot.
	w := vg.Points(20)

	var plots []plot.Plotter

	for i, v := range allValues {
		b, err := plotter.NewBoxPlot(w, float64(i), v)
		if err != nil {
			log.Fatalf("ERROR: creating box plot with %v", err)
		}
		plots = append(plots, b)
	}

	p.Add(plots...)
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = YearTicks{startYear: 2010}
	if err := p.Save(15*vg.Inch, 15*vg.Inch, "/home/markus/npm-analysis/boxplot.png"); err != nil {
		log.Fatalf("ERROR: Could not save plot with %v", err)
	}
}
