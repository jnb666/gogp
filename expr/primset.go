package expr
import (
    "strconv"
)

// A PrimSet represents the set of all of primitive opcodes for a given run. 
type PrimSet struct {
    NumVars int
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


