// Package util provides utility functions for building gogp models.
package util
import (
    "fmt"
    "os"
    "bufio"
    "flag"
    "runtime"
    "github.com/jnb666/gogp/gp"
)

// Options struct holds global configuration options
type Options struct {
    TournSize, MaxGen int
    PopSize, Threads int
    TargetFitness, CrossoverProb, MutateProb float64
    Plot, Verbose bool
    Seed int64
}

// Default option values
var DefaultOptions = Options{
    MaxGen: 40,
    PopSize: 500,
    TournSize: 7,
    TargetFitness: 0.99,
    Threads: runtime.NumCPU(),
    CrossoverProb: 0.5,
    MutateProb: 0.2,
}

// ParseFlags reads command flags and sets no. of threads and random seed.
func ParseFlags(opts *Options) {
    flag.IntVar(&opts.MaxGen, "gens", opts.MaxGen, "maximum no. of generations")
    flag.IntVar(&opts.TournSize, "tournsize", opts.TournSize, "tournament size")
    flag.IntVar(&opts.PopSize, "popsize", opts.PopSize, "population size")
    flag.IntVar(&opts.Threads, "threads", opts.Threads, "number of parallel threads")
    flag.Int64Var(&opts.Seed, "seed", opts.Seed, "random seed - set randomly if <= 0")
    flag.Float64Var(&opts.TargetFitness, "target", opts.TargetFitness, "target fitness")
    flag.Float64Var(&opts.CrossoverProb, "cxprob", opts.CrossoverProb, "crossover probability")
    flag.Float64Var(&opts.MutateProb, "mutprob", opts.MutateProb, "mutation probability")
    flag.BoolVar(&opts.Plot, "plot", opts.Plot, "serve plot data via http")
    flag.BoolVar(&opts.Verbose, "v", opts.Verbose, "print out best individual so far")
    flag.Parse()
    gp.SetSeed(opts.Seed)
    runtime.GOMAXPROCS(opts.Threads)
}

// Open opens a file for reading and returns line scanner
func Open(path string) *bufio.Scanner {
    file, err := os.Open(path)
    CheckErr(err)
    return bufio.NewScanner(file)    
}

// Read reads a line from the scanner, returns true if line was read
func Read(s *bufio.Scanner, args ...interface{}) bool {
    if s.Scan() {
        _, err := fmt.Sscan(s.Text(), args...)
        CheckErr(err)
        return true
    }
    return false
}

// CheckErr exits with an error if err is not nil
func CheckErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
