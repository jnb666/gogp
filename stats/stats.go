// Package stats provides functions for calculating, accumulating and logging statistics for gogp.
package stats
import (
    "fmt"
    "math"
    "reflect"
    "strings"
    "github.com/jnb666/gogp/gp"
)

// Formatting for logging implemented in Stats.String() method.
// The default set of columns and format strings are set on initialisation.
var (
    BrowserWidth = 850
    BrowserHeight = 950
    LogColumn = []string{"Gen", "Evals", "Fit.Max", "Fit.Avg", "Fit.Std", 
                         "Size.Avg", "Size.Max", "Depth.Avg", "Depth.Max"}
    LogFormatFloat = "%.3g"
    LogFormatInt   = "%d"
    LogColumnFormat = "%-8s"
    HistBars = 50
    Debug = false
)

// The Stats structure holds the statistics for the give Population.
type Stats struct {
    Gen, Evals int
    Fit, Size, Depth StatsData
    FitHist []int
    Best *gp.Individual
}

// The StatsData struct holds the values for a single metric.
type StatsData struct { 
    Min, Max, Avg, Std float64
    MinIndex, MaxIndex int
}

// Create calculates stats on fitness, size and depth for the given population and returns a new 
// Stats struct
func Create(pop gp.Population, gen, evals int) *Stats {
    s := &Stats{ Gen:gen, Evals:evals }
    s.Fit = updateStats(pop, func(ind *gp.Individual)float64 { return ind.Fitness })
    s.Size = updateStats(pop, func(ind *gp.Individual)float64 { return float64(ind.Size()) })
    s.Depth = updateStats(pop, func(ind *gp.Individual)float64 { return float64(ind.Depth()) })
    s.Best = pop[s.Fit.MaxIndex]
    s.FitHist = make([]int, HistBars)
    for _, ind := range pop {
        bin := int(ind.Fitness*float64(HistBars))
        if bin > HistBars-1 { bin-- }
        if bin >= 0 && bin < HistBars {
            s.FitHist[bin]++
        }
    }
    return s
}

// update stats data, calc running mean and variance
func updateStats(pop gp.Population, getval func(*gp.Individual)float64) StatsData {
    d := StatsData{ Min: 1e99, Max: 1e-99 }
    var oldM, oldS float64
    for i, ind := range pop {
        val := getval(ind)
        if val > d.Max { 
            d.Max, d.MaxIndex = val, i
        }
        if val < d.Min { 
            d.Min, d.MinIndex = val, i
        }
        if i == 0 {
            oldM, d.Avg = val, val
        } else {
            d.Avg = oldM + (val-oldM)/float64(i+1)
            d.Std = oldS + (val-oldM)*(val-d.Avg)
            oldM, oldS = d.Avg, d.Std
        }
    }
    if len(pop) > 1 {
        d.Std = math.Sqrt(d.Std / float64(len(pop)-1))
    }
    return d
}

// get struct field by reflection
func getField(struc reflect.Value, name string) (val interface{}, err error) {
    fld := struc.Elem().FieldByName(name)
    if !fld.IsValid() {
        err = fmt.Errorf("Stats field %s is not valid", name)
        return
    }
    val = fld.Interface()
    return
}

// The Get method returns the value for the named StatsData field.
func (d StatsData) Get(name string) (val interface{}, err error) {
    return getField(reflect.ValueOf(&d), name)
}

// The Get method returns the value for the named Stats field.
// If dot notation is used then it will extract the subfield from the StatsData struct.
func (s *Stats) Get(name string) (val interface{}, err error) {
    names := strings.Split(name, ".")
    if val, err = getField(reflect.ValueOf(s), names[0]); err != nil {
        return
    }
    if len(names) > 1 {
        val, err = val.(StatsData).Get(names[1])
    }
    return
}

// LogHeaders returns header names for each of the LogColumn fields
func LogHeaders() []string {
    cols := make([]string, len(LogColumn))
    for i, col := range LogColumn {
        cols[i] = strings.Replace(col, ".", "", -1)
    }
    return cols
}

// LogValues returns stats data for each of LogColumn fields
func (s *Stats) LogValues() []string {
    cols := make([]string, len(LogColumn))
    for i, col := range LogColumn {
        val, _ := s.Get(col)
        switch val.(type) {
        case float64:
            cols[i] = fmt.Sprintf(LogFormatFloat, val)
        case int:
            cols[i] = fmt.Sprintf(LogFormatInt, val)
        }
    }
    return cols
}

// String method returns formatted stats data for logging
func (s *Stats) String() string {
    cols := make([]string, len(LogColumn))
    text := ""
    if s.Gen == 0 {
        for i, col := range LogHeaders() {
            cols[i] = fmt.Sprintf(LogColumnFormat, col)
        }
        text += strings.TrimSpace(strings.Join(cols, " ")) + "\n"
    }
    for i, col := range s.LogValues() {
        cols[i] = fmt.Sprintf(LogColumnFormat, col)
    }
    // testing package does not like trailing space in examples!
    text += strings.TrimSpace(strings.Join(cols, " "))
    return text
}


