// functions to generate random expression trees
package gp
import (
    "math/rand"
	"math/big"
	crand "crypto/rand"
    "fmt"
)

// interface to generate new individuals
type Generator interface {
    Terminals()  []Opcode
    Primitives() []Opcode
    Generate() *Individual
}

type Population []*Individual

// create new population
func CreatePopulation(args *Config, gen Generator, eval Evaluator) (Population, *Stats) {
    pop := make(Population, args.PopSize)
    for i := range pop {
        pop[i] = gen.Generate()
    }
    evals := pop.Evaluate(eval, args.Threads)
    if args.Verbose { pop.Print() }
    stats := GetStats(pop, evals, nil)
    return pop, stats
}

// debug printing
func (pop Population) Print() {
    for i, ind := range pop {
        fmt.Printf("%4d: %s\n", i, *ind)
    }
}

// each generator embeds this base structure
type genBase struct { 
    *PrimSet
    min, max int
    condition func(height, depth int) bool
}

// generate an expression where each leaf has the same depth between min and max
func GenFull(pset *PrimSet, min, max int) Generator {
    return genBase{ 
        pset, min, max,
        func(height, depth int) bool {
            return depth == height
        },
    }
}

// generate an expression where each leaf may have different depth between min and max
func GenGrow(pset *PrimSet, min, max int) Generator {
    terms, prims := len(pset.Terminals()), len(pset.Primitives())
    terminalRatio := float64(terms) / float64(terms+prims)
    return genBase{
        pset, min, max,
        func(height, depth int) bool {
            return depth == height || (depth >= min && rand.Float64() < terminalRatio)
        },
    }
}

// generate an expression either with full or grow method with equal probability
type genRamped struct {
    *PrimSet
    genFull, genGrow genBase
}

func GenRamped(pset *PrimSet, min, max int) Generator {
    return genRamped{
        pset,
        GenFull(pset, min, max).(genBase),
        GenGrow(pset, min, max).(genBase),
    }
}

func (g genRamped) Generate() *Individual {
    if rand.Float64() >= 0.5 {
        return g.genGrow.Generate()
    } else {
        return g.genFull.Generate()
    }
}

// core logic which implements the different generator types
func (g genBase) Generate() *Individual {
    code := Expr{}
    height := rand.Intn(1+g.max-g.min) + g.min
    stack := []int{0}
    depth := 0
    for len(stack) > 0 {
        depth, stack = stack[len(stack)-1], stack[:len(stack)-1]
        if g.condition(height, depth) {
            op := RandomOp(g.Terminals())
            if eph,ok := op.(EphemeralConstant); ok {
                // initialise ephemeral constant
                op = eph.Init()
            }
            code = append(code, op)
        } else {
            op := RandomOp(g.Primitives())
            code = append(code, op)
            for i:=0; i<op.Arity(); i++ {
                stack = append(stack, depth+1)
            }
        }
    }
    return &Individual{Code: code}
}

// utils
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

func RandomOp(list []Opcode) Opcode {
    return list[rand.Intn(len(list))]
}


