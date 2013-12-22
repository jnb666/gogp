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
    "runtime"
    "runtime/pprof"
    "math/rand"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
)

type Config struct {
    maxGens, tournSize, maxSize, maxDepth int
    seed int64
    targetFitness float64
    datafile, profile string
}

type Point struct { x, y float64 }

// function to generate the random constant generator function
func ercGen(start, end int) func()num.V {
    fmt.Println("generate random constants in range", start, "to", end)
    return func()num.V {
        return num.V(start + rand.Intn(end-start+1))
    }
}

// returns function to calc least squares difference and return as normalised fitness from 0->1
func fitnessFunc(trainSet []Point) func(gp.Expr) (float64, bool) {
    return func(code gp.Expr) (float64, bool) {
        diff := 0.0
        for _, pt := range trainSet {
            val := float64(code.Eval(num.V(pt.x)).(num.V))
            diff += (val-pt.y)*(val-pt.y)
        }
        return 1.0/(1.0+diff), true
    }
}

// returns function to log stats for each generation
func statsLogger(maxGens int, targetFitness float64) func(*gp.Stats) bool {
    var best = gp.Individual{}
    return func(stats *gp.Stats) bool {
        fmt.Println(stats)
        // print best if fitness is improved
        if stats.Best.Fitness > best.Fitness {
            best = *stats.Best
            fmt.Println(best)
        }
        if stats.Fitness.Max >= targetFitness {
            fmt.Println("** SUCCESS **")
            return true
        }
        return stats.Generation > maxGens
    }
}

// main GP routine
func main() {
    problem := &gp.Model{}
    args := getArgs(problem)
    ercMin, ercMax, trainSet := getData(args.datafile)
    gp.SetSeed(args.seed)
	runtime.GOMAXPROCS(problem.Threads)

    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg)
    pset.Add(num.Ephemeral("ERC", ercGen(ercMin, ercMax)))

    problem.PrimitiveSet = pset
    problem.Generator = gp.GenRamped(pset, 1, 3)
    problem.Fitness = fitnessFunc(trainSet)
    problem.Offspring = gp.Tournament(args.tournSize)
    problem.Mutate = gp.MutUniform(gp.GenRamped(pset, 0, 2))
    problem.Crossover = gp.CxOnePoint()
    if args.maxDepth > 0 {
        problem.AddDecorator(gp.DepthLimit(args.maxDepth))
    }
    if args.maxSize > 0 { 
        problem.AddDecorator(gp.SizeLimit(args.maxSize))
    }
    problem.PrintParams("== GP Symbolic Regression ==")
	if args.profile != "" {
		if file, err := os.Create(args.profile); err == nil {
    		fmt.Println("writing CPU profile data to ", args.profile)
    		pprof.StartCPUProfile(file)
    		defer pprof.StopCPUProfile()
        }
	}
    problem.Run(statsLogger(args.maxGens, args.targetFitness))
}

// process cmd line flags and read input file
func getArgs(m *gp.Model) *Config {
    args := &Config{}  
	flag.IntVar(&args.maxGens, "gens", 40, "maximum no. of generations")
	flag.Float64Var(&args.targetFitness, "target", 0.99, "target fitness")
	flag.IntVar(&args.tournSize, "tournsize", 5, "tournament size")
	flag.IntVar(&args.maxSize, "size", 0, "maximum tree size - zero for none")
	flag.IntVar(&args.maxDepth, "depth", 0, "maximum tree depth - zero for none")
	flag.IntVar(&m.PopSize, "popsize", 500, "population size")
	flag.IntVar(&m.Threads, "threads", runtime.NumCPU(), "number of parallel threads")
	flag.Float64Var(&m.CrossoverProb, "cxprob", 0.5, "crossover probability")
	flag.Float64Var(&m.MutateProb, "mutprob", 0.2, "mutation probability")
	flag.Int64Var(&args.seed, "seed", 0, "random seed - set randomly if <= 0")
	flag.StringVar(&args.datafile, "trainset", "poly.dat", "file with training function")
	flag.StringVar(&args.profile, "cpuprofile", "", "write cpu profile to file")
	flag.Parse()
    return args
}

// read data file
func getData(filename string) (ERCmin, ERCmax int, trainSet []Point) {
    file, err := os.Open(filename)
    defer file.Close()
    checkErr(err)
    scanner := bufio.NewScanner(file)
    // first line has params for random constant generation
    getLine(scanner, &ERCmin, &ERCmax)
    // rest are x and y points
    trainSet = []Point{}
    var p Point
    for getLine(scanner, &p.x, &p.y) {
        trainSet = append(trainSet, p)
    }
    return
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

