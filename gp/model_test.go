// This example demonstrates using the Model type to encapsulate a problem.
package gp_test

import (
    "fmt"
    "github.com/jnb666/gogp/gp"
    . "github.com/jnb666/gogp/num"
)

// calc least squares difference and return as normalised fitness from 0->1
func getFitness(code gp.Expr) (float64, bool) {
    diff := 0.0
    for x := -1.0; x <= 1.0; x += 0.1 {
        val := float64(code.Eval([]gp.Value{V(x)}).(V))
        fun := x*x*x*x + x*x*x + x*x + x
        diff += (val-fun)*(val-fun)
    }
    return 1.0/(1.0+diff), true
}

// callback for each generation, returns true to exit the run
func logStats(s *gp.Stats) bool {
    fmt.Println(s)
    return s.Generation > 40 || s.Best.Fitness >= 1
}

func ExampleModel() {
    gp.SetSeed(1)
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(Add, Sub, Mul, Div, Neg, V(0), V(1))

    problem := gp.Model{
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
    /* Output:
set random seed: 1
gen      evals    fitMax   fitAvg   fitStd   sizeAvg  sizeMax  depthAvg depthMax
0        500      0.12     0.025    0.014    6.85     15       1.96     3
1        299      0.33     0.0344   0.0203   6.33     27       1.93     6
2        286      0.663    0.0469   0.0448   6.26     27       1.9      7
3        265      0.663    0.0598   0.0683   6.58     34       2.06     9
4        280      0.663    0.0772   0.0879   7.51     39       2.39     9
5        291      0.663    0.0918   0.1      8.92     32       2.82     8
6        302      0.663    0.117    0.133    10.3     35       3.2      10
7        294      1        0.152    0.169    11.1     35       3.48     10
    */
}



