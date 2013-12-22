// Package plot provides an interface to plot gogp data using plotinum.
package plot
import (
    "image/color"
    pplot "code.google.com/p/plotinum/plot"
    "code.google.com/p/plotinum/plotter"
    "code.google.com/p/plotinum/plotutil"
)

type Plot struct {
    *pplot.Plot
    colorIndex int
}

// The New function creates a new Plotinum plot
func New(title string) (*Plot, error) {
    p, err := pplot.New()
    if err != nil { return nil, err }
    p.Title.Text = title
    p.X.Label.Text = "generation"
	p.X.Padding, p.Y.Padding = 0, 0
    p.Legend.Top, p.Legend.Left = true, true
	p.Add(plotter.NewGrid())
    return &Plot{ Plot:p }, nil
}

// The AddLine function adds a line plot for the given statistics metric
func (p *Plot) AddLine(name string, points plotter.XYer) error {
    line, err := plotter.NewLine(points)
    if err != nil { return err }
    line.Color = p.nextColor()
    p.Add(line)
    p.Legend.Add(name, line)
    return nil
}

// The AddLinesErrors function adds line plot with Y error bars for the given statistics metric
func (p *Plot) AddLineErrors(name string, points interface{
        plotter.XYer
        plotter.YErrorer
    }) error {
    // the line
    line, err := plotter.NewLine(points)
    if err != nil { return err }
    line.Color = p.nextColor()
    p.Add(line)
    p.Legend.Add(name, line)
    // and the errors
    bars, err := plotter.NewYErrorBars(points)
    if err != nil { return err }
    bars.Color = line.Color
    p.Add(bars)
    return nil
}

func (p *Plot) nextColor() color.Color {
    col := plotutil.Color(p.colorIndex)
    p.colorIndex++
    return col
}


