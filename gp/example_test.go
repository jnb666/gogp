// This example demonstrates a simple symbolic regression problem.
package gp_test

import (
	"fmt"
	"github.com/jnb666/gogp/gp"
	"github.com/jnb666/gogp/num"
)

type eval struct{}

// calc least squares difference and return as normalised fitness from 0->1
func (e eval) GetFitness(code gp.Expr) (float64, bool) {
	diff := 0.0
	for x := -1.0; x <= 1.0; x += 0.1 {
		val := float64(code.Eval(num.V(x)).(num.V))
		fun := x*x*x*x + x*x*x + x*x + x
		diff += (val - fun) * (val - fun)
	}
	return 1.0 / (1.0 + diff), true
}

func Example_gp() {
	// create initial population
	gp.SetSeed(1)
	pset := gp.CreatePrimSet(1, "x")
	pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg, num.V(0), num.V(1))
	generator := gp.GenFull(pset, 1, 3)
	pop, evals := gp.CreatePopulation(500, generator).Evaluate(eval{}, 1)
	best := pop.Best()
	fmt.Printf("gen=%d evals=%d fit=%.4f\n", 0, evals, best.Fitness)

	// setup genetic variations
	tournament := gp.Tournament(3)
	mutate := gp.MutUniform(gp.GenGrow(pset, 0, 2))
	crossover := gp.CxOnePoint()

	// loop till reach target fitness or exceed no. of generations
	for gen := 1; gen <= 40 && best.Fitness < 1; gen++ {
		offspring := tournament.Select(pop, len(pop))
		pop, evals = gp.VarAnd(offspring, crossover, mutate, 0.5, 0.2).Evaluate(eval{}, 1)
		best = pop.Best()
		fmt.Printf("gen=%d evals=%d fit=%.4f\n", gen, evals, best.Fitness)
	}
	fmt.Println(best.Code.Format())
	// Output:
	// set random seed: 1
	// gen=0 evals=500 fit=0.1203
	// gen=1 evals=299 fit=0.3299
	// gen=2 evals=286 fit=0.6633
	// gen=3 evals=265 fit=0.6633
	// gen=4 evals=280 fit=0.6633
	// gen=5 evals=291 fit=0.6633
	// gen=6 evals=302 fit=0.6633
	// gen=7 evals=294 fit=1.0000
	// (x + (((x / 1) - ((x / 1) * -(((x * x) + x)))) * (1 * x)))
}
