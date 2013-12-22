package main
// Polynomial symbolic regression problem.
// Aim is to generate a function which will match the input dataset.
// Use trainset.py to generate this file from an arbitrary expression.

import (
    "os"
    "bufio"
    "fmt"
    "flag"
    "strings"
    "reflect"
    "runtime"
    "runtime/pprof"
    "math/rand"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
)

// Config is used for global configuration settings
type Config struct {
    Threads int
    Seed int64
    Generations int
    PopSize int
    TournamentSize int
    DepthLimit int
    SizeLimit int
    CrossoverProb float64
    MutateProb float64
    ERCmin, ERCmax int
    DataFile string
}

const TARGET = 0.99

type Point [2]float64

// function to generate the random constant generator function
func ercGen(start, end int) func()num.V {
    return func()num.V {
        return num.V(start + rand.Intn(end-start+1))
    }
}

// implement the evaluator interface to get fitness
type EvalFitness struct { 
    *gp.PrimSet 
    trainSet []Point
}

// calc least squares difference and return as normalised fitness from 0->1
func (e EvalFitness) GetFitness(code gp.Expr) (float64, bool) {
    diff := 0.0
    for _, r := range e.trainSet {
        val := float64(code.Eval([]gp.Value{num.V(r[0])}).(num.V))
        diff += (val-r[1])*(val-r[1])
    }
    return 1.0/(1.0+diff), true
}

// main GP routine
func main() {
    var best = &gp.Individual{}
    args, profile := getArgs()
    trainSet := getData(args)
    printParams(args)
    gp.SetSeed(args.Seed)
	runtime.GOMAXPROCS(args.Threads)

    // create initial population
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg)
    pset.Add(num.Ephemeral("ERC", ercGen(args.ERCmin, args.ERCmax)))
    generate := gp.GenRamped(1, 3)
    eval := EvalFitness{ pset, trainSet }
    pop, evals := gp.CreatePopulation(args.PopSize, generate, pset).Evaluate(eval, args.Threads)
    stats := gp.GetStats(pop, 0, evals)
    fmt.Println(stats)

    // setup genetic variations
    tournament := gp.Tournament(args.TournamentSize)
    mutate := gp.MutUniform(gp.GenRamped(0, 2), pset)
    crossover := gp.CxOnePoint()
    if args.DepthLimit > 0 {
        limit := gp.DepthLimit(args.DepthLimit)
        mutate.AddDecorator(limit)
        crossover.AddDecorator(limit)
    }
    if args.SizeLimit > 0 {
        limit := gp.SizeLimit(args.SizeLimit)
        mutate.AddDecorator(limit)
        crossover.AddDecorator(limit)
    }

	if profile != "" {
		if file, err := os.Create(profile); err == nil {
    		fmt.Println("writing CPU profile data to ", profile)
    		pprof.StartCPUProfile(file)
    		defer pprof.StopCPUProfile()
        }
	}

    // loop till reach target fitness or exceed no. of generations   
    for gen := 1; gen <= args.Generations; gen++ {
        if stats.Fitness.Max >= TARGET {
            fmt.Println("** SUCCESS **")
            break
        }
        offspring := tournament.Select(pop, args.PopSize)
        pop, evals = gp.VarAnd(offspring, crossover, mutate, 
                        args.CrossoverProb, args.MutateProb).Evaluate(eval, args.Threads)
        stats = gp.GetStats(pop, gen, evals)
        fmt.Println(stats)
        if stats.Best.Fitness > best.Fitness {
            best = stats.Best
            fmt.Println(best)
        }
    }
}

// process cmd line flags and read input file
func getArgs() (args *Config, profile string) {
    args = &Config{}
	flag.IntVar(&args.Threads, "threads", runtime.NumCPU(), "number of parallel threads")
	flag.Int64Var(&args.Seed, "seed", 0, "random seed - set randomly if <= 0")
	flag.IntVar(&args.Generations, "gens", 40, "maximum no. of generations")
	flag.IntVar(&args.PopSize, "popsize", 500, "population size")
	flag.IntVar(&args.TournamentSize, "tournsize", 5, "tournament size")
	flag.IntVar(&args.SizeLimit, "size", 0, "maximum tree size - zero for none")
	flag.IntVar(&args.DepthLimit, "depth", 0, "maximum tree depth - zero for none")
	flag.Float64Var(&args.CrossoverProb, "cxprob", 0.5, "crossover probability")
	flag.Float64Var(&args.MutateProb, "mutprob", 0.2, "mutation probability")
	flag.StringVar(&profile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&args.DataFile, "trainset", "poly.dat", "file with training function")
	flag.Parse()
    return
}

// print the config parameters for this run
func printParams(args *Config) {
    fmt.Printf("== GP Symbolic Regression ==\n")
	s := reflect.ValueOf(args).Elem()
    for i:=0; i<s.NumField(); i++ {
		fmt.Printf("%14s = %v\n", s.Type().Field(i).Name, s.Field(i).Interface())
    }
}

// read data file
func getData(args *Config) []Point {
    file, err := os.Open(args.DataFile)
    defer file.Close()
    checkErr(err)
    scanner := bufio.NewScanner(file)
    getLine(scanner, &args.ERCmin, &args.ERCmax)
    trainSet := []Point{}
    var p Point
    for getLine(scanner, &p[0], &p[1]) {
        trainSet = append(trainSet, p)
    }
    return trainSet
}

func getLine(s *bufio.Scanner, item1, item2 interface{}) bool {
    if !s.Scan() { return false }
    items := strings.Split(s.Text(), "\t")
    _, err := fmt.Sscan(items[0], item1)
    checkErr(err)
    _, err = fmt.Sscan(items[1], item2)
    checkErr(err)
    return true
}

func checkErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

