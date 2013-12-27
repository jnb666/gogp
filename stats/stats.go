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
    LogFormatFloat = "%.3g"
    LogFormatInt   = "%d"
    LogColumnFormat = "%-8s"
)

// Stats structure holds the statistics for the give Population.
// Struct tags are used for descriptive field name in JSON encoding
type Stats struct {
    Gen     int             `desc:"generation"`
    Evals   int             `desc:"no. of evals"`
    Fit     StatsData       `desc:"fitness"`
    Size    StatsData       `desc:"size"`
    Depth   StatsData       `desc:"depth"`
    Done    bool
}

// Stats data holds the values for a single metric.
type StatsData struct { 
    Min  float64    `desc:"minimum"`
    Max  float64    `desc:"maximum"`
    Avg  float64    `desc:"mean"`
    Std  float64    `desc:"std deviation"`
    MinIndex, MaxIndex int
}

// StatsHistory slice stores all the stats for a given run
type StatsHistory []*Stats

// Plot struct is used for encoding the StatsHistory to JSON for plotting using flot
type Plot struct {
    Label string        `json:"label"`
    Lines struct {
        Fill bool       `json:"fill"`
        LineWidth int   `json:"lineWidth"`
    }                   `json:"lines"`
    Data  [][3]float64  `json:"data"`
}

// Create calculates stats on fitness, size and depth for the given population and returns a new 
// Stats struct
func Create(pop gp.Population, gen, evals int) *Stats {
    s := &Stats{ Gen:gen, Evals:evals }
    s.Fit = updateStats(pop, func(ind *gp.Individual)float64 { return ind.Fitness })
    s.Size = updateStats(pop, func(ind *gp.Individual)float64 { return float64(ind.Size()) })
    s.Depth = updateStats(pop, func(ind *gp.Individual)float64 { return float64(ind.Depth()) })
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

// get struct field and desc tag by reflection
func getField(struc reflect.Value, name string) (val interface{}, tag string, err error) {
    fld := struc.Elem().FieldByName(name)
    if !fld.IsValid() {
        err = fmt.Errorf("Stats field %s is not valid", name)
        return
    }
    val = fld.Interface()
    if tfld, ok := struc.Elem().Type().FieldByName(name); ok {
        tag = tfld.Tag.Get("desc")
    }
    return
}

// The Get method returns the data and desc tag for the named StatsData field,
// or an error if the field does not exist.
func (d StatsData) Get(name string) (val interface{}, tag string, err error) {
    return getField(reflect.ValueOf(&d), name)
}

// The Get method returns the data and desc tag for the named field or an error if field does not exist. 
// If dot notation is used then it will extract the subfield from the StatsData struct.
func (s *Stats) Get(name string) (val interface{}, tag string, err error) {
    names := strings.Split(name, ".")
    if val, tag, err = getField(reflect.ValueOf(s), names[0]); err != nil {
        return
    }
    if len(names) > 1 {
        val, tag, err = val.(StatsData).Get(names[1])
    }
    return
}

// The Get method returns the history data for the named field in a suitable format for plotting
// Returns FieldNotFound error if name is not a valid field
func (h StatsHistory) Get(name string) (lines []Plot, err error) {
    var val interface{}
    lines = make([]Plot, 3)
    for i, field := range []string{"Max", "Avg", "Std"} {
        lines[i].Data = make([][3]float64, len(h))
        if field == "Std" {
            lines[i].Lines.Fill = true
            lines[i].Lines.LineWidth = 0
        } else {
            lines[i].Lines.LineWidth = 2
        }
        for j, stats := range h {
            if val, lines[i].Label, err = stats.Get(name + "." + field); err != nil {
                return
            }
            if y, ok := val.(float64); ok {
                if field == "Std" {
                    avg := lines[i-1].Data[j][1]
                    lines[i].Data[j] = [3]float64{ float64(j), avg-y, avg+y }
                } else {
                    lines[i].Data[j] = [3]float64{ float64(j), y, 0 }
                }
            } else {
                err = fmt.Errorf("Stats field %s could not be converted to float", name)
            }
        }
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
        val, _,  _ := s.Get(col)
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


