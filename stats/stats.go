// Package stats provides functions for calculating, accumulating and logging statistics for gogp.
package stats
import (
    "fmt"
    "math"
    "reflect"
    "strings"
    "encoding/json"
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
// Struct tags are used for descriptive field name in JSON encoding
type Stats struct {
    Gen     int             `desc:"generation"`
    Evals   int             `desc:"no. of evals"`
    Best    *gp.Individual  `desc:"best individual"`
    Fit     *StatsData      `desc:"fitness"`
    Size    *StatsData      `desc:"size"`
    Depth   *StatsData      `desc:"depth"`
}

// Stats data holds the values for a single metric.
type StatsData struct { 
    Min  float64    `desc:"minimum"`
    Max  float64    `desc:"maximum"`
    Avg  float64    `desc:"mean"`
    Std  float64    `desc:"std deviation"`
}

// StatsHistory slice stores all the stats for a given run
type StatsHistory []*Stats

// PlotData struct is used for encoding the StatsHistory to JSON e.g. for plotting
type PlotData struct {
    Label string        `json:"label"`
    Data  [][2]float64  `json:"data"`
}

// GetStats calculates stats on fitness, size and depth for the given population
func Create(pop gp.Population, gen, evals int) *Stats {
    s := &Stats{ Gen:gen, Evals:evals, Fit:NewStatsData(), Size:NewStatsData(), Depth:NewStatsData() }
    updateStats(pop, s.Fit, func(ind *gp.Individual)float64 { return ind.Fitness })
    updateStats(pop, s.Size, func(ind *gp.Individual)float64 { return float64(ind.Size()) })
    updateStats(pop, s.Depth, func(ind *gp.Individual)float64 { return float64(ind.Depth()) })
    s.Best = pop.Best().Clone()
    return s
}

// NewStatsData initialises a new StatsData struct
func NewStatsData() *StatsData {
    return &StatsData{ Min: 1e99, Max: 1e-99 }
}

// update stats data, calc running mean and variance
func updateStats(pop gp.Population, d *StatsData, getval func(*gp.Individual)float64) {
    var oldM, oldS float64
    for i, ind := range pop {
        val := getval(ind)
        if val > d.Max { d.Max = val }
        if val < d.Min { d.Min = val }
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
func (d *StatsData) Get(name string) (val interface{}, tag string, err error) {
    return getField(reflect.ValueOf(d), name)
}

// The Get method returns the data and desc tag for the named field or an error if field does not exist. 
// If dot notation is used then it will extract the subfield from the StatsData struct.
func (s *Stats) Get(name string) (val interface{}, tag string, err error) {
    names := strings.Split(name, ".")
    if val, tag, err = getField(reflect.ValueOf(s), names[0]); err != nil {
        return
    }
    if len(names) > 1 {
        val, tag, err = val.(*StatsData).Get(names[1])
    }
    return
}

// The Get method returns the history data for the named field in a suitable format for plotting
// Returns FieldNotFound error if name is not a valid field
func (h StatsHistory) Get(name string) (p PlotData, err error) {
    var val interface{}
    p.Data  = make([][2]float64, len(h))
    p.Label = name
    for i, stats := range h {
        if val, p.Label, err = stats.Get(name); err != nil {
            return
        }
        if fval, ok := val.(float64); ok { 
            p.Data[i] = [2]float64{ float64(i), fval }
        } else {
            err = fmt.Errorf("Stats field %s could not be converted to float", name)
        }
    }
    return
}

// The GetJSON method calls Get to retrieve the data, then encodes in JSON format for plotting.
// name provided should be the name of a StatsData struct (Fit / Size / Depth)
func (h StatsHistory) GetJSON(name string) ([]byte, error) {
    data, err := h.Get(name)
    if err != nil { return nil, err }
    return json.Marshal(data)
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
        val, _, _ := s.Get(col)
        cols[i] = fmt.Sprintf(format, val)
    }
    // testing package does not like trailing space in examples!
    text += strings.TrimSpace(strings.Join(cols, " "))
    return text
}


