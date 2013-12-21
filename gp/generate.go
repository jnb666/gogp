package gp
import (
    "math/rand"
	"math/big"
	crand "crypto/rand"
    "fmt"
    "github.com/jnb666/gogp/expr"
)

// A PrimSet represents the set of all of primitive opcodes for a given run. 
type PrimSet struct {
    terminals  []expr.Opcode
    primitives []expr.Opcode
}


// CreatePrimitiveSet constructor takes a list of variable names.
func CreatePrimSet(vars ...string) *PrimSet {
    pset := &PrimSet{}
    pset.primitives = []expr.Opcode{}
    pset.terminals = make([]expr.Opcode, len(vars))
    for i, name := range vars {
        pset.terminals[i] = expr.Variable(name, i)
    }
    return pset
}

// Add method adds a new Opcode to the primitive set.
func (pset *PrimSet) Add(ops ...expr.Opcode) {
    for _, op := range ops {
        if op.Arity() > 0 {
            pset.primitives = append(pset.primitives, op)
        } else {
            pset.terminals = append(pset.terminals, op)
        }
    }
}

// Var returns the nth variable in the primitive set.
func (pset *PrimSet) Var(n int) expr.Opcode {
    return pset.terminals[n]
}

// Terminals returns the list of terminal arity 0 opcodes. 
func (pset *PrimSet) Terminals() []expr.Opcode {
    return pset.terminals
}

// Primitives returns the list of non-terminal opcodes (arity > 0).
func (pset *PrimSet) Primitives() []expr.Opcode { 
    return pset.primitives
}

// A Population is a slice of individuals. Implementations of the Selection interface are 
// provided to pick a subset from the population, and of the Variation interface to provide mutation
// and crossover genetic operators. Decorators can be used to further customise a Variation, 
// e.g. for size limits for bloat control.
type Population []*Individual

// CreatePopulation creates a new population of args.PopSize using the provided generator and
// primitive set. The evaluator is called to calculate the fitness for each individual and 
// stats are calculated for the new population.
func CreatePopulation(args *Config, gen Generator, pset *PrimSet, eval Evaluator) (Population, *Stats) {
    pop := make(Population, args.PopSize)
    for i := range pop {
        pop[i] = gen.Generate(pset)
    }
    evals := pop.Evaluate(eval, args.Threads)
    if args.Verbose { pop.Print() }
    stats := GetStats(pop, evals, nil)
    return pop, stats
}

// Print prints out each member of the population.
func (pop Population) Print() {
    for i, ind := range pop {
        fmt.Printf("%4d: %s\n", i, *ind)
    }
}

// A Generator is used to generate new individuals from the provided primitive set, 
// typically by a random expression generation algorithm. 
type Generator interface {
    Generate(*PrimSet) *Individual
}

// each generator embeds this base structure
type genBase struct {
    min, max int
    condition func(height, depth int, pset *PrimSet) bool
}

// GenFull returns a generator to produce individuals with expression trees such
// that each leaf has the same depth between min and max.
func GenFull(min, max int) Generator {
    return genBase{ 
        min, max,
        func(height, depth int, pset *PrimSet) bool {
            return depth == height
        },
    }
}

// GenGrow returns a generator to produce individuals with expression trees such
// that each leaf may have different depth between min and max.
func GenGrow(min, max int) Generator {
    return genBase{
        min, max,
        func(height, depth int, pset *PrimSet) bool {
            terms, prims := len(pset.Terminals()), len(pset.Primitives())
            terminalRatio := float64(terms) / float64(terms+prims)
            return depth == height || (depth >= min && rand.Float64() < terminalRatio)
        },
    }
}

type genRamped struct {
    full, grow genBase
}

// GenRamped returns a generator which uses either the GenFull or GenRamped algorithm
// with equal probability.
func GenRamped(min, max int) Generator {
    return genRamped{
        GenFull(min, max).(genBase),
        GenGrow(min, max).(genBase),
    }
}

func (g genRamped) Generate(pset *PrimSet) *Individual {
    if rand.Float64() >= 0.5 {
        return g.grow.Generate(pset)
    } else {
        return g.full.Generate(pset)
    }
}

// core logic which implements the different generator types
func (g genBase) Generate(pset *PrimSet) *Individual {
    code := expr.Expr{}
    height := rand.Intn(1+g.max-g.min) + g.min
    stack := []int{0}
    depth := 0
    for len(stack) > 0 {
        depth, stack = stack[len(stack)-1], stack[:len(stack)-1]
        if g.condition(height, depth, pset) {
            op := randomOp(pset.Terminals())
            if erc,ok := op.(expr.EphemeralConstant); ok {
                op = erc.Init()
            }
            code = append(code, op)
        } else {
            op := randomOp(pset.Primitives())
            code = append(code, op)
            for i:=0; i<op.Arity(); i++ {
                stack = append(stack, depth+1)
            }
        }
    }
    return &Individual{Code: code}
}

// SetSeed sets the random number seed to seed, or to a random value if seed is <= 0
func SetSeed(seed int64) int64 {
	if seed <= 0 {
		max := big.NewInt(2<<31-1)
		rseed, _ := crand.Int(crand.Reader, max)
		seed = rseed.Int64()
	}
    fmt.Println("set random seed:", seed)
    rand.Seed(seed)
    return seed
}

func randomOp(list []expr.Opcode) expr.Opcode {
    return list[rand.Intn(len(list))]
}


