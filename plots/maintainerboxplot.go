package plots

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"math/rand"
)

func boxPlotExample() {
	// Get some data to display in our plot.
	rand.Seed(int64(0))
	n := 1000000
	uniform := make(plotter.Values, n)
	normal := make(plotter.Values, n)
	expon := make(plotter.Values, n)
	for i := 0; i < n; i++ {
		uniform[i] = rand.Float64() * 100.0
		normal[i] = rand.NormFloat64() * 100.0
		expon[i] = rand.ExpFloat64() * 100.0
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
	if err := p.Save(8*vg.Inch, 9*vg.Inch, "boxplot.png"); err != nil {
		panic(err)
	}
}
