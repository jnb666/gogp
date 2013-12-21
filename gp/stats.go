package gp
import (
    "fmt"
    "math"
)

// Columns for statistics logging 
var LogColumns = []string{"gen", "evals", "fitMax", "fitAvg", "fitStd",
                     "sizeAvg", "sizeMax", "depthAvg", "depthMax"}

// Stats structure holds the statistics for the give Population. 
type Stats struct {
    Generation int
    Evals   int
    Best    *Individual
    Fitness StatsData
    Size    StatsData
    Depth   StatsData
}

type StatsData struct {
    Min, Max, Avg, Std float64
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

// String method returns formatted stats data for logging
func (s *Stats) String() string {
    text := ""
    if s.Generation == 0 {
        for _, col := range LogColumns {
            text += fmt.Sprintf("%-8s ", col)
        }
        text += "\n"
    }
    text += fmt.Sprintf("%-8d %-8d %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g", 
        s.Generation, s.Evals, s.Fitness.Max, s.Fitness.Avg, s.Fitness.Std, 
        s.Size.Avg, s.Size.Max, s.Depth.Avg, s.Depth.Max)
    return text
}


