// stats info for the current generation
package gp
import (
    "fmt"
    "math"
)

type Stats struct {
    Evals   int
    Fitness statsData
    Size    statsData
    Depth   statsData
    Best    Individual
    NewBest bool
}

type statsData struct {
    Min, Max, Avg, Std float64
    Imin, Imax int
}

var LogColumns = []string{"gen", "evals", "fitMax", "fitAvg", "fitStd",
                     "sizeAvg", "sizeMax", "depthAvg", "depthMax"}

// get stats for this generation
func GetStats(pop Population, evals int, prev *Stats) *Stats {
    s := &Stats{ Evals:evals }
    updateStats(pop, &s.Fitness, func(ind *Individual)float64 { return ind.Fitness })
    updateStats(pop, &s.Size, func(ind *Individual)float64 { return float64(ind.Size()) })
    updateStats(pop, &s.Depth, func(ind *Individual)float64 { return float64(ind.Depth()) })
    s.Best = *(pop[s.Fitness.Imax].Clone())
    s.NewBest = prev == nil || !Equals(&s.Best, &prev.Best)
    return s
}

// update stats data
func updateStats(pop Population, d *statsData, getval func(*Individual)float64) {
    psize := float64(len(pop))
    d.Min, d.Max = 1e99, -1e99
    for i, ind := range pop {
        val := getval(ind)
        if val > d.Max { d.Max,d.Imax = val,i }
        if val < d.Min { d.Min,d.Imin = val,i }
        d.Avg += val / psize
    }
    for _, ind := range pop {
        val := getval(ind)
        d.Std += (val-d.Avg)*(val-d.Avg) / psize
    }
    d.Std = math.Sqrt(d.Std)
}

// print current stats
func (s *Stats) Print(gen int) {
    if gen == 0 {
        for _, col := range LogColumns {
            fmt.Printf("%-8s ", col)
        }
        fmt.Println()
    }
    fmt.Printf("%-8d %-8d %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g %-8.3g\n", 
        gen, s.Evals, s.Fitness.Max, s.Fitness.Avg, s.Fitness.Std, 
        s.Size.Avg, s.Size.Max, s.Depth.Avg, s.Depth.Max)
    if s.NewBest {
        fmt.Println(s.Best.Code.Format())
    }
}






