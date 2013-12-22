// Package plot provides an interface to plot gogp data using plotinum.
package plot
import (
    "image/color"
    pplot "code.google.com/p/plotinum/plot"
    "code.google.com/p/plotinum/plotter"
    "code.google.com/p/plotinum/plotutil"
    "github.com/jnb666/gogp/gp"
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
func (p *Plot) AddLine(model *gp.Model, field string) error {
    line, err := plotter.NewLine(model.GetHistory(field))
    if err != nil { return err }
    line.Color = p.nextColor()
    p.Add(line)
    p.Legend.Add(field, line)
    return nil
}

// The AddLinesErrors function adds line plot with Y error bars for the given statistics metric
func (p *Plot) AddLineErrors(model *gp.Model, field, errField string) error {
    data := model.GetHistoryErrors(field, errField)
    // the line
    line, err := plotter.NewLine(data)
    if err != nil { return err }
    line.Color = p.nextColor()
    p.Add(line)
    p.Legend.Add(field, line)
    // and the errors
    bars, err := plotter.NewYErrorBars(data)
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


