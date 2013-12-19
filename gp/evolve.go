// evolve gp population
package gp
import (
    "math/rand"
    _ "fmt"
)

// global configuration settings
type Config struct {
    Verbose bool
    Threads int
    Seed int64
    Generations int
    PopSize int
    TournamentSize int
    DepthLimit int
    CrossoverProb float64
    MutateProb float64
}

// interface for selecting individuals from population, should use clone to make a deep copy
type Selector interface {
    Select(pop Population, num int) Population
}

// decorator interface - can add a list off functions which will be applied on each individual
type Decorator interface {
    Decorate(in, out *Individual) *Individual
}

// interface for applying crossover or mutation to a code tree
type Variation interface {
    AddDecorator(Decorator)
    Variate(ind Population) Population
}

// select new individuals, apply variations and evaluate fitness
func NextGeneration(args *Config, pop Population, sel Selector, cx, mut Variation,
        eval Evaluator, stats *Stats) (Population, *Stats) {
    pop = sel.Select(pop, len(pop))
    pop = VarAnd(pop, cx, mut, args.CrossoverProb, args.MutateProb)
    evals := pop.Evaluate(eval, args.Threads)
    if args.Verbose { pop.Print() }
    stats = GetStats(pop, evals, stats)
    return pop, stats
}

// apply crossover and mutation variations
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

// deep copy of individuals in population
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
}

func (v *variation) AddDecorator(decor Decorator) {
    v.decorators = append(v.decorators, decor)
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

// uniform mutation - select random point in tree and replace with generated expression
func MutUniform(gen Generator) *variation {
    mutate := func(ind Population) Population {
        tree := ind[0].Code
        pos := rand.Intn(len(tree))
        newtree := gen.Generate().Code
        ind[0] = NewIndiv(replaceSubtree(pos, tree, newtree))
        return ind
    }
    return &variation{ []Decorator{}, mutate }
}

// select random point in each individual and exchange subtrees between the two trees 
func CxOnePoint() *variation {
    cross := func(ind Population) Population {
        if ind[0].Size() < 2 || ind[1].Size() < 2 {
            return ind
        }
        pos1, subtree1 := getRandomSubtree(ind[0].Code)
        pos2, subtree2 := getRandomSubtree(ind[1].Code)
        ind[0] = NewIndiv(replaceSubtree(pos1, ind[0].Code, subtree2))
        ind[1] = NewIndiv(replaceSubtree(pos2, ind[1].Code, subtree1))
        return ind
    }
    return &variation{ []Decorator{}, cross }
}

// replace subtree in code at pos with subtree - code is updated, but subtree arg is not
func replaceSubtree(pos int, code, subtree Expr) Expr {
    end := code.Traverse(pos, nil, nil)
    tail := subtree
    if end < len(code)-1 {
        tail = append(tail.Clone(), code[end+1:]...)
    }
    return append(code[:pos], tail...)
}

// return a copy of nodes in randomly selected subtree of code - returns a copy
func getRandomSubtree(code Expr) (pos int, subtree Expr) {
    pos = rand.Intn(len(code))
    end := code.Traverse(pos, nil, nil)
    subtree = code[pos:end+1].Clone()
    return
}

// get best individual by fitness
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
type Tournament struct { TournamentSize int }

func (s Tournament) Select(pop Population, num int) Population {
    chosen := Population{}
    for i := 0; i < num; i++ {
        group := RandomSel{}.Select(pop, s.TournamentSize)
        best := group.Best()
        if !best.FitnessValid {
            panic("no best individual found!")
        }
        chosen = append(chosen, best)
	}
    return chosen
}

// random selection - select random samples from population
type RandomSel struct { }

func (s RandomSel) Select(pop Population, num int) Population {
    chosen := Population{}
    for i := 0; i < num; i++ {
        chosen = append(chosen, pop[rand.Intn(len(pop))])
    }
    return chosen
}

// decorators which we can apply to variations
type SizeLimit struct { Max int }

func (d SizeLimit) Decorate(in, out *Individual) *Individual {
    if out.Size() > d.Max {
        return in
    } else {
        return out
    }
}

type DepthLimit struct { Max int }

func (d DepthLimit) Decorate(in, out *Individual) *Individual {
    if out.Depth() > d.Max {
        return in
    } else {
        return out
    }
}

    


