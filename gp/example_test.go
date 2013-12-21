// This example demonstrates a simple symbolic regression problem.
package gp_test

import (
    "fmt"
    "github.com/jnb666/gogp/expr"
    "github.com/jnb666/gogp/gp"
    . "github.com/jnb666/gogp/num"
)

type EvalFitness struct { *expr.PrimSet }

// calc least squares difference and return as normalised fitness from 0->1
func (e EvalFitness) GetFitness(code expr.Expr) (float64, bool) {
    diff := 0.0
    for x := -1.0; x <= 1.0; x += 0.1 {
        val := float64(code.Eval([]expr.Value{V(x)}).(V))
        fun := x*x*x*x + x*x*x + x*x + x
        diff += (val-fun)*(val-fun)
    }
    return 1.0/(1.0+diff), true
}

func Example_gp() {
    // create initial population
    gp.SetSeed(1)
    pset := expr.CreatePrimSet(1, "x")
    pset.Add(Add, Sub, Mul, Div, Neg, V(0), V(1))
    eval := EvalFitness{ pset }
    pop, evals := gp.CreatePopulation(500, gp.GenFull(1,3), pset).Evaluate(eval, 1)
    best := pop.Best()
    fmt.Printf("gen=%d evals=%d fit=%.4f\n", 0, evals, best.Fitness)

    // setup genetic variations
    tourn  := gp.Tournament(3)
    mutate := gp.MutUniform(gp.GenGrow(0,2), pset)
    cxover := gp.CxOnePoint()

    // loop till reach target fitness or exceed no. of generations   
    for gen := 1; gen <= 40 && best.Fitness < 1; gen++ {
        offspring := tourn.Select(pop, len(pop))
        pop, evals = gp.VarAnd(offspring, cxover, mutate, 0.5, 0.2).Evaluate(eval, 1)
        best = pop.Best()
        fmt.Printf("gen=%d evals=%d fit=%.4f\n", gen, evals, best.Fitness)
    }
    fmt.Println(best)
    /* Output:
set random seed: 1
gen=0 evals=500 fit=0.1203
gen=1 evals=299 fit=0.3299
gen=2 evals=286 fit=0.6633
gen=3 evals=265 fit=0.6633
gen=4 evals=280 fit=0.6633
gen=5 evals=291 fit=0.6633
gen=6 evals=302 fit=0.6633
gen=7 evals=294 fit=1.0000
 1.000  (x + (((x / 1) - ((x / 1) * -(((x * x) + x)))) * (1 * x)))
*/
}





