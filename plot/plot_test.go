// This example demonstrates plotting a graph of the fitness scores.
package plot

import (
    "testing"
    "fmt"
    "os"
    "os/exec"
    "path"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
)

var PLOT_FILE string

func init() {
    PLOT_FILE = path.Join(os.TempDir(), "gogp-fitness.png")
}

// calc least squares difference and return as normalised fitness from 0->1
func getFitness(code gp.Expr) (float64, bool) {
    diff := 0.0
    for x := -1.0; x <= 1.0; x += 0.1 {
        val := float64(code.Eval(num.V(x)).(num.V))
        fun := x*x*x*x + 2*x*x*x + 3*x*x + x - 1
        diff += (val-fun)*(val-fun)
    }
    return 1.0/(1.0+diff), true
}

// callback for each generation, returns true to exit the run
func logStats(s *gp.Stats) bool {
    fmt.Println(s)
    return s.Generation > 40 || s.Best.Fitness >= 1
}

// run linear regression and generate a fitness plot
func TestPlot(t *testing.T) {
    gp.SetSeed(0)
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg, num.V(0), num.V(1))

    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 1, 3),
        PopSize: 500,
        Fitness: getFitness,
        Offspring: gp.Tournament(3),
        Mutate: gp.MutUniform(gp.GenGrow(pset, 0, 2)),
        MutateProb: 0.2,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: 0.5,
        Threads: 1,
    }
    problem.Run(logStats)

    // generate a plot of fitness vs generation
    t.Log("plot test")
    p, err := New("symolic regression example")
    checkError(t, err)
    err = p.AddLine("max fitness", gp.GetStatsHistory(problem, "fitMax", ""))
    checkError(t, err)
    err = p.AddLineErrors("mean fitness", gp.GetStatsHistory(problem, "fitAvg", "fitStd"))
    checkError(t, err)
    t.Log("save plot to", PLOT_FILE)
    err = p.Save(6, 6, PLOT_FILE)
    checkError(t, err)
    t.Log("try and show it with display")
    err = exec.Command("display", PLOT_FILE).Run()
    checkError(t, err)
}

func checkError(t *testing.T, e error) {
    if e != nil { t.Fatal(e) }
}

