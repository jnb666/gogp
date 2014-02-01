package main

// Boolean even parity problem
// aim is to generate a function which will return the even parity bit for PARITY_FANIN boolean inputs

import (
	"fmt"
	"github.com/jnb666/gogp/boolean"
	"github.com/jnb666/gogp/gp"
	"github.com/jnb666/gogp/stats"
	"github.com/jnb666/gogp/util"
	"math"
)

const PARITY_FANIN = 6
const FORMAT = "%06b"
const TARGET = 0.99

// check each of the 2**PARITY_FANIN cases to get parity at initialisation time
func getFitnessFunc() func(gp.Expr) (float64, bool) {
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
		return float64(correct) / float64(paritySize), true
	}
}

// main GP routine
func main() {
	opts := util.DefaultOptions
	util.ParseFlags(&opts)

	pset := gp.CreatePrimSet(PARITY_FANIN)
	pset.Add(boolean.And, boolean.Or, boolean.Xor, boolean.Not, boolean.True, boolean.False)

	problem := &gp.Model{
		PrimitiveSet:  pset,
		Generator:     gp.GenFull(pset, 3, 5),
		PopSize:       opts.PopSize,
		Fitness:       getFitnessFunc(),
		Offspring:     gp.Tournament(opts.TournSize),
		Mutate:        gp.MutUniform(gp.GenGrow(pset, 0, 2)),
		MutateProb:    opts.MutateProb,
		Crossover:     gp.CxOnePoint(),
		CrossoverProb: opts.CrossoverProb,
		Threads:       opts.Threads,
	}
	problem.PrintParams("== Even parity problem for", PARITY_FANIN, "inputs ==")

	logger := stats.NewLogger(opts.MaxGen, opts.TargetFitness)
	if opts.Plot {
		stats.MainLoop(problem, logger, ":8080", "../web")
	} else {
		fmt.Println()
		logger.PrintStats = true
		logger.PrintBest = opts.Verbose
		problem.Run(logger)
	}
}
