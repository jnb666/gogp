package main
// Boolean even parity problem
// aim is to generate a function which will return the even parity bit for PARITY_FANIN boolean inputs

import (
    "fmt"
    "math"
    "flag"
    "time"
    "runtime"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
    "github.com/jnb666/gogp/boolean"
)

const PARITY_FANIN = 6
const FORMAT = "%06b"
const TARGET = 0.99

// check each of the 2**PARITY_FANIN cases to get parity at initialisation time
func getFitnessFunc() func(gp.Expr) (float64,bool) {
    paritySize := int(math.Pow(2, PARITY_FANIN))
    input := make([][]gp.Value, paritySize)
    output := make([]gp.Value, paritySize)
    for i := range output {
        input[i] = make([]gp.Value, PARITY_FANIN)
        bitstr := fmt.Sprintf(FORMAT, i)
        parity := true
        for j, bit := range bitstr {
            if bit == '1' {
                input[i][j] = boolean.True
                parity = !parity
            } else {
                input[i][j] = boolean.False
            }
        }
        output[i] = boolean.V(parity)
    }
    // fitness is no. of correct cases / total
    return func(code gp.Expr) (float64, bool) {
        correct := 0
        for i, in := range input {
            if code.Eval(in...) == output[i] {
                correct++
            }
        }
        return float64(correct)/float64(paritySize) , true
    }
}

// main GP routine
func main() {
    var threads, generations, popsize int
    var seed int64
    var plot, verbose bool
	flag.IntVar(&threads, "threads", runtime.NumCPU(), "number of parallel threads")
	flag.Int64Var(&seed, "seed", 0, "random seed - set randomly if <= 0")
	flag.IntVar(&generations, "gens", 40, "maximum no. of generations")
	flag.IntVar(&popsize, "popsize", 1000, "population size")
	flag.BoolVar(&plot, "plot", false, "connect to gogpweb to plot statistics")
	flag.BoolVar(&verbose, "v", false, "print out best individual so far")
    flag.Parse()

    pset := gp.CreatePrimSet(PARITY_FANIN)
    pset.Add(boolean.And, boolean.Or, boolean.Xor, boolean.Not, boolean.True, boolean.False)

    problem := gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 3, 5),
        PopSize: popsize,
        Fitness: getFitnessFunc(),
        Offspring: gp.Tournament(3),
        Mutate: gp.MutUniform(gp.GenGrow(pset, 0, 2)),
        MutateProb: 0.2,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: 0.5,
        Threads: threads,
    }
    problem.PrintParams("== Even parity problem for", PARITY_FANIN, "inputs ==")
    gp.SetSeed(seed)
	runtime.GOMAXPROCS(threads)
    fmt.Println()

    logger := &stats.Logger{
        MaxGen: generations, 
        TargetFitness: TARGET,
        PrintStats: true,
        PrintBest: verbose,
    }
    if plot {
        go logger.ListenAndServe(":8080", "../web")
        stats.StartBrowser("http://localhost:8080")
    }
    problem.Run(logger)
    if plot {
        time.Sleep(1*time.Hour)
    }
}


