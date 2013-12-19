// interface to evaluate fitness of individual
package gp
import "fmt"

type Evaluator interface {
    GetFitness(code Expr) (fit float64, ok bool)
}

type Individual struct {
    Code  Expr
    Fitness  float64
    FitnessValid bool
    depth int
}

type empty struct{}

// work out fitness for each individual, split work into threads parallel goroutines
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
        //fmt.Println("doEval:", start, end)
        go doEval(pop, eval, todo[start:end], sem)
        start += chunkSize
        end += chunkSize
    }
    // wait for goroutines to finish
    for chunk := 0; chunk < threads; chunk++ { <-sem }
    return evals
}

func doEval(pop Population, eval Evaluator, todo []int, sem chan empty) {
    for _, i := range todo {
         pop[i].Fitness, pop[i].FitnessValid = eval.GetFitness(pop[i].Code)
    }
    sem <- empty{};
}

// individual methods
func NewIndiv(code Expr) *Individual {
    return &Individual{ Code:code.Clone() }
}

func (ind *Individual) Clone() *Individual {
    return &Individual{
        Code: ind.Code.Clone(),
        Fitness: ind.Fitness,
        FitnessValid: ind.FitnessValid,
    }
}

func (ind Individual) String() string {
    if ind.FitnessValid {
        return fmt.Sprintf("%6.3f  %s", ind.Fitness, ind.Code.Format())
    } else {
        return fmt.Sprintf("%6s  %s", "????", ind.Code.Format())
    }
}

func (ind *Individual) Size() int {
    return len(ind.Code)
}

func (ind *Individual) Depth() int {
    if len(ind.Code) == 0 {
        return 0
    }
    if ind.depth == 0 {
        ind.depth = ind.Code.Depth()
    }
    return ind.depth
}

func Equals(ind1, ind2 *Individual) bool {
    return ind1.Code.Format() == ind2.Code.Format()
}



