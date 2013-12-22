package gp
import (
    "fmt"
    "math"
    "reflect"
    "strings"
)

// Formatting for logging implemented in Stats.String() method.
// The default set of columns and format strings are set on initialisation 
var (
    LogColumn = []string{"gen", "evals", "fitMax", "fitAvg", "fitStd", "sizeAvg", "sizeMax", "depthAvg", "depthMax"}
    LogColumnFmt = []string{"d", "d", ".3g", ".3g", ".3g", ".3g", ".3g", ".3g", ".3g"}
    LogColumnWidth = 8
)

// Stats structure holds the statistics for the give Population. 
type Stats struct {
    Generation int
    Evals   int
    Best    *Individual
    Fitness StatsData
    Size    StatsData
    Depth   StatsData
}

// Stats data holds the values for a single metric
type StatsData struct { Min, Max, Avg, Std float64 }

// StatsHistory implements the plotinum plotter.XYer and plotter.YErrorer interfaces
type StatsHistory []struct{X, Y, Err float64}

// Len returns the number of elements in the slice of points
func (h StatsHistory) Len() int { return len(h) }

// XY returns the ith point in the slice
func (h StatsHistory) XY(i int) (x, y float64) {
    x, y = h[i].X, h[i].Y
    return
}

// Yerror returns the high and low errors
func (h StatsHistory) YError(i int) (low, high float64) {
    low, high = h[i].Err, h[i].Err
    return
}

// GetStats calculates stats on fitness, size and depth for the given population
func GetStats(pop Population, gen, evals int) *Stats {
    s := &Stats{ Generation:gen, Evals:evals }
    updateStats(pop, &s.Fitness, func(ind *Individual)float64 { return ind.Fitness })
    updateStats(pop, &s.Size, func(ind *Individual)float64 { return float64(ind.Size()) })
    updateStats(pop, &s.Depth, func(ind *Individual)float64 { return float64(ind.Depth()) })
    s.Best = pop.Best().Clone()
    return s
}

// update stats data
func updateStats(pop Population, d *StatsData, getval func(*Individual)float64) {
    psize := float64(len(pop))
    d.Min, d.Max = 1e99, -1e99
    for _, ind := range pop {
        val := getval(ind)
        if val > d.Max { d.Max = val }
        if val < d.Min { d.Min = val }
        d.Avg += val / psize
    }
    for _, ind := range pop {
        val := getval(ind)
        d.Std += (val-d.Avg)*(val-d.Avg) / psize
    }
    d.Std = math.Sqrt(d.Std)
}

// The Get method returns the data in the named field, name must be a valid field else this will panic
func (d *StatsData) Get(name string) interface{} {
    fld := reflect.ValueOf(d).Elem().FieldByName(name)
    return fld.Interface()
}

// The Get method returns the data in the named field where field name is one of
// the field names defined in LogColumn package variable
func (s *Stats) Get(name string) (val interface{}) {
    switch {
        case name == "gen":
            val = s.Generation
        case name == "evals":
            val = s.Evals
        case strings.HasPrefix(name, "fit"):
            val = s.Fitness.Get(strings.TrimPrefix(name, "fit"))
        case strings.HasPrefix(name, "size"):
            val = s.Size.Get(strings.TrimPrefix(name, "size"))
        case strings.HasPrefix(name, "depth"):
            val = s.Depth.Get(strings.TrimPrefix(name, "depth"))
        default:
            panic(name + " does not reference any valid StatsData field")
    }
    return
}

// String method returns formatted stats data for logging
func (s *Stats) String() string {
    cols := make([]string, len(LogColumn))
    text := ""
    if s.Generation == 0 {
        format := fmt.Sprintf("%%-%ds", LogColumnWidth) 
        for i, col := range LogColumn {
            cols[i] = fmt.Sprintf(format, col)
        }
        text += strings.TrimSpace(strings.Join(cols, " ")) + "\n"
    }
    for i, col := range LogColumn {
        format := fmt.Sprintf("%%-%d%s", LogColumnWidth, LogColumnFmt[i])
        cols[i] = fmt.Sprintf(format, s.Get(col))
    }
    // testing package does not like trailing space in examples!
    text += strings.TrimSpace(strings.Join(cols, " "))
    return text
}


