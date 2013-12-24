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
// The default set of columns and format strings are set on initialisation 
var (
    LogColumn = []string{"Gen", "Evals", "Fit.Max", "Fit.Avg", "Fit.Std", 
                         "Size.Avg", "Size.Max", "Depth.Avg", "Depth.Max"}
    LogColumnFmt = []string{"d", "d", ".3g", ".3g", ".3g", ".3g", ".3g", ".3g", ".3g"}
    LogColumnWidth = 8
)

// Stats structure holds the statistics for the give Population. 
type Stats struct {
    Gen     int
    Evals   int
    Best    *gp.Individual
    Fit     StatsData
    Size    StatsData
    Depth   StatsData
}

// Stats data holds the values for a single metric
type StatsData struct { Min, Max, Avg, Std float64 }

// GetStats calculates stats on fitness, size and depth for the given population
func Create(pop gp.Population, gen, evals int) *Stats {
    s := &Stats{ Gen:gen, Evals:evals }
    updateStats(pop, &s.Fit, func(ind *gp.Individual)float64 { return ind.Fitness })
    updateStats(pop, &s.Size, func(ind *gp.Individual)float64 { return float64(ind.Size()) })
    updateStats(pop, &s.Depth, func(ind *gp.Individual)float64 { return float64(ind.Depth()) })
    s.Best = pop.Best().Clone()
    return s
}

// update stats data
func updateStats(pop gp.Population, d *StatsData, getval func(*gp.Individual)float64) {
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

// The Get method returns the data in the named field. If dot notation is used then it will
// extract the subfield from the StatsData struct
func (s *Stats) Get(name string) (val interface{}) {
    el := reflect.ValueOf(s).Elem()
    if fields := strings.Split(name, "."); len(fields) > 1 {
        data := el.FieldByName(fields[0]).Interface().(StatsData)
        return data.Get(fields[1])
    }
    return el.FieldByName(name).Interface()
}

// String method returns formatted stats data for logging
func (s *Stats) String() string {
    cols := make([]string, len(LogColumn))
    text := ""
    if s.Gen == 0 {
        format := fmt.Sprintf("%%-%ds", LogColumnWidth) 
        for i, col := range LogColumn {
            col = strings.Replace(col, ".", "", -1)
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


