package gp
import (
    "fmt"
)

// An Evaluator is provided by the implementation to calculate the fitness of an individual.
// The fitness should be a normalised fitness value, i.e. a number in the range 0 to 1 
// where 0 is the worst possible and 1 represents a perfect solution to the problem.
type Evaluator interface {
    GetFitness(code Expr) (fit float64, ok bool)
}

// An Individual element of the population has a code expression which represents the genome 
// and a fitness value as calculated by the implementation of the Evaluator interface.
// Methods are provided to apply generic operations to individuals via the Variator interface.
type Individual struct {
    Code  Expr
    Fitness  float64
    FitnessValid bool
    depth int
}

type empty struct{}

// Evaluate calls the eval Evaluator to calculate the fitness for each individual.
// Work can be split into threads parallel goroutines.
func (pop Population) Evaluate(eval Evaluator, threads int) int {
    todo := make([]int, 0, len(pop))
    for i, ind := range pop {
        if !ind.FitnessValid { todo = append(todo, i) }
    }
    // split work into threads chunks
    evals := len(todo)
    chunkSize := evals/threads
    if chunkSize < 1 { chunkSize = 1 }
    start := 0
    end := chunkSize
    sem := make(chan empty, threads)
    for chunk := 0; chunk < threads; chunk++ {
        // last chunk takes any extras
        if chunk == threads-1 { end = evals }
        // kick off goroutine to do the work
        go func(indices []int, sem chan empty) {
            for _, i := range indices {
                 pop[i].Fitness, pop[i].FitnessValid = eval.GetFitness(pop[i].Code)
            }
            sem <- empty{}
        } (todo[start:end], sem)
        start += chunkSize
        end += chunkSize
    }
    // wait for goroutines to finish
    for chunk := 0; chunk < threads; chunk++ { <-sem }
    return evals
}

// NewIndiv constructor, to create a new individual with copy of given code tree.
func NewIndiv(code Expr) *Individual {
    return &Individual{ Code:code.Clone() }
}

// Clone returns a copy of the given individual.
func (ind *Individual) Clone() *Individual {
    return &Individual{
        Code: ind.Code.Clone(),
        Fitness: ind.Fitness,
        FitnessValid: ind.FitnessValid,
    }
}

// String returns a textual representation of the individual, e.g. for debug printing.
func (ind Individual) String() string {
    if ind.FitnessValid {
        return fmt.Sprintf("%6.3f  %s", ind.Fitness, ind.Code.Format())
    } else {
        return fmt.Sprintf("%6s  %s", "????", ind.Code.Format())
    }
}

// Size returns the length of the code for the individual.
func (ind *Individual) Size() int {
    return len(ind.Code)
}

// Depth returns the depth of the code tree for the individual.
func (ind *Individual) Depth() int {
    if len(ind.Code) == 0 {
        return 0
    }
    if ind.depth == 0 {
        ind.depth = ind.Code.Depth()
    }
    return ind.depth
}

// Equals compares two individuals to see if they have the same code value.
func Equals(ind1, ind2 *Individual) bool {
    return ind1.Code.Format() == ind2.Code.Format()
}



