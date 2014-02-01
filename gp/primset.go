package gp

import (
	"fmt"
	"strconv"
)

// A PrimSet represents the set of all of primitive opcodes for a given run.
// NumVars is the number of input variables, Terminals a list of all the terminal zero arity nodes
// and Primitives are the nodes which have one or more arguments.
type PrimSet struct {
	NumVars    int
	Terminals  []Opcode
	Primitives []Opcode
}

// CreatePrimitiveSet constructs a new primitive set with nvars input variables.
// Names for the variables can optionally be specified in varNames, else will default to inn
func CreatePrimSet(nvars int, varNames ...string) *PrimSet {
	pset := &PrimSet{}
	pset.NumVars = nvars
	pset.Primitives = []Opcode{}
	pset.Terminals = make([]Opcode, nvars)
	for i := 0; i < nvars; i++ {
		var name string
		if len(varNames) > i {
			name = varNames[i]
		} else {
			name = "in" + strconv.Itoa(i)
		}
		pset.Terminals[i] = Variable(name, i)
	}
	return pset
}

// String returns a string representation of the list of primitives
func (pset *PrimSet) String() string {
	var ops Expr
	ops = append(pset.Terminals, pset.Primitives...)
	return fmt.Sprint(ops)
}

// Add method adds a new Opcode to the primitive set.
func (pset *PrimSet) Add(ops ...Opcode) {
	for _, op := range ops {
		if op.Arity() > 0 {
			pset.Primitives = append(pset.Primitives, op)
		} else {
			pset.Terminals = append(pset.Terminals, op)
		}
	}
}

// Var returns the nth variable in the primitive set.
func (pset *PrimSet) Var(n int) Opcode {
	return pset.Terminals[n]
}
