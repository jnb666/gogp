package main

import (
    "math"
    "math/rand"
    "runtime"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
    "github.com/jnb666/gogp/stats"
)

// calc least squares difference and return as normalised fitness from 0->1
func getFitness(code gp.Expr) (float64, bool) {
    diff := 0.0
    for x := 0.0; x <= 6.2; x += 0.1 {
        val := float64(code.Eval(num.V(x)).(num.V))
        fun := math.Sin(x)
        diff += (val-fun)*(val-fun)
    }
    return 1.0/(1.0+diff), true
}

// set up problem
func createModel() (*gp.Model, *Data) {
    gp.SetSeed(0)
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg)
    pset.Add(num.Ephemeral("ERC", func()num.V { return num.V(-5 + rand.Intn(11)) }))

    problem := gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenRamped(pset, 1, 3),
        PopSize: 20000,
        Fitness: getFitness,
        Offspring: gp.Tournament(5),
        Mutate: gp.MutUniform(gp.GenRamped(pset, 0, 2)),
        MutateProb: 0.2,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: 0.5,
        Threads: 4,
    }
    problem.AddDecorator(gp.SizeLimit(500))
	runtime.GOMAXPROCS(4)
    problem.PrintParams("== GP Symbolic Regression ==")

    history := Data{ 
        Stats:   stats.StatsHistory{},
        MaxGens: 40,
    }
    return &problem, &history
}

// goroutine to run the model - updates history struct
func runModel(problem *gp.Model, history *Data) {
    problem.Run(
        func(pop gp.Population, gen, evals int) bool {
            s := stats.Create(pop, gen, evals)
            history.Lock()
            defer history.Unlock()
            history.Stats = append(history.Stats, s)
            history.Done  = (gen >= history.MaxGens || s.Best.Fitness >= 1)
            return history.Done
        })
}


