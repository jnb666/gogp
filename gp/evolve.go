// gogp is a library for Koza style genetic programming in Go.
package gp
import (
    "math/rand"
    "fmt"
    "reflect"
)

// interface for selecting individuals from population, should use clone to make a deep copy
type Selector interface {
    Select(pop Population, num int) Population
    String() string
}

// Variation is an interface for applying a genetic operation to one or more Individuals.
type Variation interface {
    AddDecorator(Decorator)
    Variate(ind Population) Population
    String() string
}

// Decorator is an interface for providing a function which can wrap a call to a Variation.
type Decorator interface {
    Decorate(in, out *Individual) *Individual
    String() string
}

// The Model type encapsulates a complete problem
type Model struct {
    PrimitiveSet *PrimSet
    PopSize, Threads int
    Generator Generator
    Offspring Selector
    MutateProb, CrossoverProb float64
    Mutate, Crossover Variation
    Fitness func(Expr) (float64, bool)
}

// The Logger interface is used for logging stats on each generation of a run
type Logger interface {
    Log(pop Population, gen, evals int) bool
}

// The GetFitness method is provided so that the Model type implements the Evaluator interface
func (m *Model) GetFitness(code Expr) (float64, bool) {
    return m.Fitness(code)
}

// AddDecorator method adds a decorator function to the mutate and crossover operations
func (m *Model) AddDecorator(decor Decorator) { 
    m.Mutate.AddDecorator(decor)
    m.Crossover.AddDecorator(decor)
}

// The Run method first creates a new population and iteratively evolves it
// using the VarAnd algorithm. The Log method is called on the Logger for each generation. 
// If it returns true then the run terminates.
func (m *Model) Run(l Logger) Population {
    gen, evals := 0, 0
    pop := CreatePopulation(m.PopSize, m.Generator)
    pop, evals = pop.Evaluate(m, m.Threads)
    for !l.Log(pop, gen, evals) {
        gen++
        offspring := m.Offspring.Select(pop, m.PopSize)
        pop = VarAnd(offspring, m.Crossover, m.Mutate, m.CrossoverProb, m.MutateProb)
        pop, evals = pop.Evaluate(m, m.Threads)
    }
    return pop
}

// PrintParams prints the config parameters for this run to stdout
func (m *Model) PrintParams(title ...interface{}) {
    fmt.Println(title...)
	s := reflect.ValueOf(m).Elem()
    for i:=0; i<s.NumField()-1; i++ {
		fmt.Printf("%14s = %v\n", s.Type().Field(i).Name, s.Field(i).Interface())
    }
}

// VarAnd is a simple algorith to apply crossover and mutation variations with given probabilities.
func VarAnd(pop Population, cross, mutate Variation, cx_prob, mut_prob float64) Population {
    offspring := pop.Clone()
    for i:=1; i<len(pop); i+=2 {
        if rand.Float64() < cx_prob {
            children := cross.Variate(offspring[i-1:i+1])
            offspring[i-1], offspring[i] = children[0], children[1]
        }
    }
    for i:=0; i<len(pop); i++ {
        if rand.Float64() < mut_prob {
            children := mutate.Variate(offspring[i:i+1])
            offspring[i] = children[0]
        }
    }
    return offspring
}

// Clone makes a deep copy of all of the individuals in the population.
func (pop Population) Clone() Population {
    newpop := make(Population, len(pop))
    for i, ind := range pop {
        newpop[i] = ind.Clone()
    }
    return newpop
}

// mutators and breeders embed this base variation type
type variation struct {
    decorators []Decorator
    vfunc func(in Population) (out Population)
    name string
}

func (v *variation) String() string {
    return v.name
}

func (v *variation) AddDecorator(decor Decorator) {
    v.decorators = append(v.decorators, decor)
    v.name += fmt.Sprintf("<%s>", decor)
}

func (v *variation) Variate(in Population) Population {
    out := v.vfunc(in.Clone())
    for _, decor := range v.decorators {
        for i := range in {
            out[i] = decor.Decorate(in[i], out[i])
        }
    }
    return out
}

// MutUniform returns a mutation variation which operates on an Individual.
// A random point in the code tree is selected and is replaced by a tree generated by the
// provided Generator from the pset primitive set.
func MutUniform(gen Generator) Variation {
    mutate := func(ind Population) Population {
        tree := ind[0].Code
        pos := rand.Intn(len(tree))
        newtree := gen.Generate().Code
        ind[0] = Create(tree.ReplaceSubtree(pos, newtree))
        return ind
    }
    return &variation{ []Decorator{}, mutate, fmt.Sprintf("MutUniform(%s)", gen) }
}

// CxOnePoint returns a crossover Variation which operates on a pair of Individuals.
// A random point in each individual is selected subtrees exchanged between the two trees. 
func CxOnePoint() Variation {
    cross := func(ind Population) Population {
        if ind[0].Size() < 2 || ind[1].Size() < 2 {
            return ind
        }
        pos1, subtree1 := ind[0].Code.RandomSubtree()
        pos2, subtree2 := ind[1].Code.RandomSubtree()
        ind[0] = Create(ind[0].Code.ReplaceSubtree(pos1, subtree2))
        ind[1] = Create(ind[1].Code.ReplaceSubtree(pos2, subtree1))
        return ind
    }
    return &variation{ []Decorator{}, cross, "CxOnePoint" }
}

// Best returns the best individual by fitness.
func (pop Population) Best() *Individual {
    best := &Individual{}
    for _, ind := range pop {
        if ind.FitnessValid && (!best.FitnessValid || ind.Fitness > best.Fitness) {
            best = ind
        }
    }
    return best
}

// tournament selection - select best out of TournamentSize random samples
type tournament struct { TournamentSize int }

// Tournament returns a selector to select the the best out of tsize random samples from the population.
func Tournament(tsize int) Selector {
    return tournament{ tsize }
}

func (s tournament) String() string {
    return fmt.Sprintf("Tournament(%d)", s.TournamentSize)
}

func (s tournament) Select(pop Population, num int) Population {
    chosen := Population{}
    for i := 0; i < num; i++ {
        group := randomSel{}.Select(pop, s.TournamentSize)
        best := group.Best()
        if !best.FitnessValid {
            panic("no best individual found!")
        }
        chosen = append(chosen, best)
	}
    return chosen
}

type randomSel struct { }

// RandomSel returns a selector to select random samples from population.
func RandomSel() Selector {
    return randomSel{}
}

func (s randomSel) String() string {
    return "RandomSel"
}

func (s randomSel) Select(pop Population, num int) Population {
    chosen := Population{}
    for i := 0; i < num; i++ {
        chosen = append(chosen, pop[rand.Intn(len(pop))])
    }
    return chosen
}

type decorBase struct { name string }

func (d decorBase) String() string { return d.name }

// SizeLimit returns a decorator which applies a static limit to the maximum expression size. 
func SizeLimit(max int) Decorator {
    return sizeLimit{ decorBase{fmt.Sprintf("SizeLimit(%d)",max)}, max }
}

type sizeLimit struct {
    decorBase
    max int
}

// If the maxium size is exceeded, return the individual prior to variation.
func (d sizeLimit) Decorate(in, out *Individual) *Individual {
    if out.Size() > d.max {
        return in
    } else {
        return out
    }
}

// DepthLimit returns a decorator which applies a static limit to the maximum tree depth.
func DepthLimit(max int) Decorator {
    return depthLimit{ decorBase{fmt.Sprintf("DepthLimit(%d)",max)}, max }
}

type depthLimit struct {
    decorBase
    max int
}

// If the maxium depth is exceeded, return the individual prior to variation.
func (d depthLimit) Decorate(in, out *Individual) *Individual {
    if out.Depth() > d.max {
        return in
    } else {
        return out
    }
}


    


