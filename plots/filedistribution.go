package plots

import (
	"github.com/markuszm/npm-analysis/util"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func PlotFiledistributionAllPackages(values []util.Pair, filePath string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = "Filetype distribution"
	p.Y.Label.Text = "Count"

	w := vg.Points(20)

	var bars []plot.Plotter
	var names []string

	for i, p := range values {
		bar, err := plotter.NewBarChart(plotter.Values{float64(p.Value)}, w)
		if err != nil {
			return err
		}
		bar.LineStyle.Width = vg.Length(0)
		bar.Color = plotutil.Color(i)

		if i < len(values)/2 {
			bar.Offset = vg.Points(-float64(len(values)/2-i) * w.Points())
		} else if i > len(values)/2 {
			bar.Offset = vg.Points(float64(i-len(values)/2) * w.Points())
		}

		bars = append(bars, bar)
		names = append(names, p.Key)
	}

	p.Add(bars...)
	p.NominalX(names...)

	if err := p.Save(5*vg.Inch, 3*vg.Inch, filePath); err != nil {
		return err
	}

	return nil
}
