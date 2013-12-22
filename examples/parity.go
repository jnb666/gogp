package main
// Boolean even parity problem
// aim is to generate a function which will return the even parity bit for PARITY_FANIN boolean inputs

import (
    "fmt"
    "math"
    "flag"
    "runtime"
    "github.com/jnb666/gogp/gp"
    . "github.com/jnb666/gogp/boolean"
)

const PARITY_FANIN = 6
const FORMAT = "%06b"
const TARGET = 0.99

// implements gp.Evaluator
type EvalFitness struct { 
    *gp.PrimSet
    size int
    in  [][]gp.Value
    out []gp.Value
}

// check each of the 2**PARITY_FANIN cases to get parity at initialisation time
func Fitness() EvalFitness {
    paritySize := int(math.Pow(2, PARITY_FANIN))
    input := make([][]gp.Value, paritySize)
    output := make([]gp.Value, paritySize)
    for i := range output {
        input[i] = make([]gp.Value, PARITY_FANIN)
        bitstr := fmt.Sprintf(FORMAT, i)
        parity := true
        for j, bit := range bitstr {
            if bit == '1' {
                input[i][j] = True
                parity = !parity
            } else {
                input[i][j] = False
            }
        }
        output[i] = V(parity)
    }
    return EvalFitness{ size:paritySize, in:input, out:output }
}

// fitness is no. of correct cases / total
func (e EvalFitness) GetFitness(code gp.Expr) (float64, bool) {
    correct := 0
    for i, input := range e.in {
        res := code.Eval(input)
        if res == e.out[i] { correct++ }
    }
    return float64(correct)/float64(e.size) , true
}

// main GP routine
func main() {
    var threads, generations, popsize int
    var seed int64
	flag.IntVar(&threads, "threads", runtime.NumCPU(), "number of parallel threads")
	flag.Int64Var(&seed, "seed", 0, "random seed - set randomly if <= 0")
	flag.IntVar(&generations, "gens", 40, "maximum no. of generations")
	flag.IntVar(&popsize, "popsize", 1000, "population size")
    flag.Parse()
    gp.SetSeed(seed)
    fmt.Println("no. of parallel threads is ", threads)
	runtime.GOMAXPROCS(threads)

    // create initial generation
    pset := gp.CreatePrimSet(PARITY_FANIN)
    pset.Add(And, Or, Xor, Not, True, False)
    generate := gp.GenFull(3, 5)
    eval := Fitness()
    pop, evals := gp.CreatePopulation(popsize, generate, pset).Evaluate(eval, threads)
    stats := gp.GetStats(pop, 0, evals)
    fmt.Println(stats)

    // loop till reach target fitness or exceed no. of generations
    tournament := gp.Tournament(3)
    mutate := gp.MutUniform(gp.GenGrow(0, 2), pset)
    crossover := gp.CxOnePoint()
    for gen := 1; gen <= generations; gen++ {
        if stats.Fitness.Max >= TARGET {
            fmt.Println("** SUCCESS **")
            break
        }
        offspring := tournament.Select(pop, popsize)
        pop, evals = gp.VarAnd(offspring, crossover, mutate, 0.5, 0.2).Evaluate(eval, threads)
        stats = gp.GetStats(pop, gen, evals)
        fmt.Println(stats)
    }
    fmt.Println(pop.Best())
}

