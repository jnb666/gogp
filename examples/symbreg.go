package main
// Polynomial symbolic regression problem.
// Aim is to generate a function which will match the input dataset.
// Use trainset.py to generate this file from an arbitrary expression.

import (
    "fmt"
    "flag"
    "math/rand"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
    "github.com/jnb666/gogp/num"
    "github.com/jnb666/gogp/util"
)

type Point struct { x, y float64 }

// read data file
func getData(filename string) (ERCmin, ERCmax int, trainSet []Point) {
    s := util.Open(filename)
    util.Read(s, &ERCmin, &ERCmax)
    trainSet = []Point{}
    var p Point
    for util.Read(s, &p.x, &p.y) {
        trainSet = append(trainSet, p)
    }
    return
}

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

// function to plot target curve 
func plotTarget(trainSet []Point) func(gp.Population) stats.Plot {
    return func(pop gp.Population) stats.Plot {
        plot := stats.NewPlot("Target", len(trainSet))
        plot.Color = "#00ff00"
        for i, pt := range trainSet {
            plot.Data[i][0], plot.Data[i][1] = pt.x, pt.y
        }
        return plot
    }
}

// function to plot best individual
func plotBest(trainSet []Point) func(gp.Population) stats.Plot {
    return func(pop gp.Population) stats.Plot {
        plot := stats.NewPlot("Best", len(trainSet))
        plot.Color = "#ff0000"
        code := pop.Best().Code
        for i, pt := range trainSet {
            plot.Data[i][0] = pt.x
            plot.Data[i][1] = float64(code.Eval(num.V(pt.x)).(num.V))
        }
        return plot
    }
}

// main GP routine
func main() {
    // get options
    var maxSize, maxDepth int
    var dataFile string
    flag.IntVar(&maxSize, "size", 0, "maximum tree size - zero for none")
    flag.IntVar(&maxDepth, "depth", 0, "maximum tree depth - zero for none")
    flag.StringVar(&dataFile, "trainset", "poly.dat", "file with training function")
    opts := util.DefaultOptions
    util.ParseFlags(&opts)

    // create primitive set
    ercMin, ercMax, trainSet := getData(dataFile)
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div)
    pset.Add(num.Ephemeral("ERC", ercGen(ercMin, ercMax)))

    // setup model
    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenRamped(pset, 1, 3),
        PopSize: opts.PopSize,
        Fitness: fitnessFunc(trainSet),
        Offspring: gp.Tournament(opts.TournSize),
        Mutate: gp.MutUniform(gp.GenGrow(pset, 0, 2)),
        MutateProb: opts.MutateProb,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: opts.CrossoverProb,
        Threads: opts.Threads,
    }
    if maxDepth > 0 {
        problem.AddDecorator(gp.DepthLimit(maxDepth))
    }
    if maxSize > 0 { 
        problem.AddDecorator(gp.SizeLimit(maxSize))
    }
    problem.PrintParams("== GP Symbolic Regression for ", dataFile, "==")

    // run
    logger := stats.NewLogger(opts.MaxGen, opts.TargetFitness)
    if opts.Plot {
        gp.GraphDPI = "60"
        logger.RegisterPlot("graph", plotTarget(trainSet), plotBest(trainSet))
        stats.MainLoop(problem, logger, ":8080", "../web")
    } else {
        fmt.Println()
        logger.PrintStats = true
        logger.PrintBest = opts.Verbose
        problem.Run(logger)
    }
}

