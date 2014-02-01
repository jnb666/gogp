package gp

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
)

// A Population is a slice of individuals. Implementations of the Selection interface are
// provided to pick a subset from the population, and of the Variation interface to provide mutation
// and crossover genetic operators. Decorators can be used to further customise a Variation,
// e.g. for size limits for bloat control.
type Population []*Individual

// CreatePopulation creates a new population of popsize individuals using the provided generator.
func CreatePopulation(popsize int, gen Generator) Population {
	pop := make(Population, popsize)
	for i := range pop {
		pop[i] = gen.Generate()
	}
	return pop
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
	Generate() *Individual
	String() string
}

// each generator embeds this base structure
type genBase struct {
	pset      *PrimSet
	min, max  int
	condition func(height, depth int) bool
	name      string
}

func (g genBase) String() string {
	return g.name
}

// GenFull returns a generator to produce individuals with expression trees such
// that each leaf has the same depth between min and max.
func GenFull(pset *PrimSet, min, max int) Generator {
	return genBase{
		pset, min, max,
		func(height, depth int) bool { return depth == height },
		fmt.Sprintf("GenFull(%d,%d)", min, max),
	}
}

// GenGrow returns a generator to produce individuals with expression trees such
// that each leaf may have different depth between min and max.
func GenGrow(pset *PrimSet, min, max int) Generator {
	terms, prims := len(pset.Terminals), len(pset.Primitives)
	terminalRatio := float64(terms) / float64(terms+prims)
	return genBase{
		pset, min, max,
		func(height, depth int) bool {
			return depth == height || (depth >= min && rand.Float64() < terminalRatio)
		},
		fmt.Sprintf("GenGrow(%d,%d)", min, max),
	}
}

type genRamped struct {
	full, grow Generator
}

func (g genRamped) String() string {
	return fmt.Sprintf("GenRamped(%d,%d)", g.full.(genBase).min, g.full.(genBase).max)
}

// GenRamped returns a generator which uses either the GenFull or GenRamped algorithm
// with equal probability.
func GenRamped(pset *PrimSet, min, max int) Generator {
	return genRamped{
		GenFull(pset, min, max),
		GenGrow(pset, min, max),
	}
}

func (g genRamped) Generate() *Individual {
	if rand.Float64() >= 0.5 {
		return g.grow.Generate()
	} else {
		return g.full.Generate()
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
			op := randomOp(g.pset.Terminals)
			if erc, ok := op.(EphemeralConstant); ok {
				op = erc.Init()
			}
			code = append(code, op)
		} else {
			op := randomOp(g.pset.Primitives)
			code = append(code, op)
			for i := 0; i < op.Arity(); i++ {
				stack = append(stack, depth+1)
			}
		}
	}
	return &Individual{Code: code}
}

func randomOp(list []Opcode) Opcode {
	return list[rand.Intn(len(list))]
}

// SetSeed sets the random number seed to seed, or to a random value if seed is <= 0
func SetSeed(seed int64) int64 {
	if seed <= 0 {
		max := big.NewInt(2<<31 - 1)
		rseed, _ := crand.Int(crand.Reader, max)
		seed = rseed.Int64()
	}
	fmt.Println("set random seed:", seed)
	rand.Seed(seed)
	return seed
}
