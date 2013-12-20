// simple symbolic regression test
package main

import (
    "os"
    "fmt"
    "flag"
    "runtime"
    "runtime/pprof"
    _ "math"
    "math/rand"
    "github.com/jnb666/gogp/gp"
    . "github.com/jnb666/gogp/num"
)

// constants
const TARGET = 0.99
const RANGE_STEP = 0.1
const RANGE_MIN = -1.0
const RANGE_MAX = 1.0

//const RANGE_MIN = 0.0
//const RANGE_MAX = 6.2

func TargetFunc(x float64) float64 {
     return x*x*x*x + x*x*x + x*x + x
//   return math.Sin(x)
}

// terminal function to generate random integer in range -1:+1
var ercgen = Terminal("rnd", func()Num { return Num(rand.Intn(3)-1) })
//var ercgen = Terminal("rnd", func()Num { return Num(rand.Intn(11)-5) })

// implement the evaluator interface to get fitness
// least squares difference - return as normalised fitness from 0->1
type EvalFitness struct { *gp.PrimSet }

func (e EvalFitness) GetFitness(code gp.Expr) (float64, bool) {
    diff := 0.0
    for x := RANGE_MIN; x <= RANGE_MAX; x += RANGE_STEP {
        d1 := float64(code.Eval([]gp.Value{Num(x)}).(Num))
        d2 := TargetFunc(x)
        diff += (d1-d2)*(d1-d2)
    }
    return 1.0/(1.0+diff), true
}

func primSet() *gp.PrimSet {
    pset := gp.CreatePrimitiveSet("x")
    pset.Add(Add, Sub, Mul, Div, Neg)
    pset.Add(gp.Ephemeral("ERC", ercgen))
    return pset
}

// main GP routine
func main() {
    args, profile := getArgs()
    fmt.Printf("== GP Symbolic Regression ==\n")
    gp.SetSeed(args.Seed)
	runtime.GOMAXPROCS(args.Threads)
    pset := primSet()
    eval := EvalFitness{ pset }
    tournament := gp.Tournament(args.TournamentSize)
    generate := gp.GenRamped(1,3)
    mutate := gp.MutUniform(gp.GenRamped(0,2), pset)
    crossover := gp.CxOnePoint()
    if args.DepthLimit > 0 {
        limit := gp.DepthLimit(args.DepthLimit)
        mutate.AddDecorator(limit)
        crossover.AddDecorator(limit)
    }
    pop, stats := gp.CreatePopulation(args, generate, pset, eval)
    stats.Print(0)
	if profile != "" {
		if file, err := os.Create(profile); err == nil {
    		fmt.Println("writing CPU profile data to ", profile)
    		pprof.StartCPUProfile(file)
    		defer pprof.StopCPUProfile()
        }
	}
    for gen := 1; gen <= args.Generations; gen++ {
        if stats.Fitness.Max >= TARGET {
            fmt.Println("** SUCCESS **")
            break
        }
        pop, stats = gp.NextGeneration(args, pop, tournament, crossover, mutate, eval, stats)
        stats.Print(gen)
    }
}

// process cmd line flags
func getArgs() (args *gp.Config, profile string) {
    args = &gp.Config{}
	flag.BoolVar(&args.Verbose, "v", false, "verbose logging")
	flag.IntVar(&args.Threads, "threads", 1, "number of parallel threads")
	flag.Int64Var(&args.Seed, "seed", 0, "random seed - set randomly if <= 0")
	flag.IntVar(&args.Generations, "gens", 40, "maximum no. of generations")
	flag.IntVar(&args.PopSize, "popsize", 500, "population size")
	flag.IntVar(&args.TournamentSize, "tournsize", 5, "tournament size")
	flag.IntVar(&args.DepthLimit, "depth", 0, "maximum tree depth - zero for none")
	flag.Float64Var(&args.CrossoverProb, "cxprob", 0.5, "crossover probability")
	flag.Float64Var(&args.MutateProb, "mutprob", 0.2, "mutation probability")
	flag.StringVar(&profile, "cpuprofile", "", "write cpu profile to file")
	flag.Parse()
    return
}



