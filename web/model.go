package main

import (
    "fmt"
    "math"
    "math/rand"
    "runtime"
    "net"
    "encoding/gob"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
    "github.com/jnb666/gogp/stats"
)

var TCPPort = ":5555"

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
func createModel() *gp.Model {
    gp.SetSeed(0)
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg)
    pset.Add(num.Ephemeral("ERC", func()num.V { return num.V(-5 + rand.Intn(11)) }))

    problem := gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenRamped(pset, 1, 3),
        MaxGen: 20,
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
    return &problem
}

// run the model and write stats to socket on each generation
func main() {
    fmt.Println("start client");
    conn, err := net.Dial("tcp", "localhost" + TCPPort)
    if err != nil {
        fmt.Println("Connection error", err)
        return
    }
    encoder := gob.NewEncoder(conn)
    problem := createModel()

    problem.Run(
        func(pop gp.Population, gen, evals int) bool {
            s := stats.Create(pop, gen, evals)
            fmt.Println(s)
            s.Done = (gen >= problem.MaxGen || s.Fit.Max >= 1)
            encoder.Encode(s)
            return s.Done
        })

    conn.Close()
    fmt.Println("done")
}


