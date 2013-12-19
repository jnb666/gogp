// represent an expression as a list of opcodes
package gp
import (
    "strings"
    "fmt"
)

// interface types
type Value interface{}

type Opcode interface {
    Arity() int
    Eval(...Value) Value
    String() string
    Format(...string) string
}

type Expr []Opcode

type PrimSet struct {
    terminals  []Opcode
    primitives []Opcode
    inputs     []Value
}

// base func - this is embedded in concrete types
type baseFunc struct {
    name  string
    arity int
}

func (f *baseFunc) Arity() int { return f.arity }

func (f *baseFunc) String() string { return f.name }

func (f *baseFunc) Eval(args ...Value) Value { panic("abstract method!") }

// terminal is a function with zero args
type Term struct { *baseFunc }

func Terminal(name string) Term {
    return Term{ &baseFunc{name, 0} }
}

func (t Term) Format(args ...string) string { return t.name + "()" }

// a function with zero or more args
type Func struct { *baseFunc }

func Function(name string, arity int) Func {
    return Func{ &baseFunc{name, arity} }
}

func (f Func) Format(args ...string) string {
    return f.name + "(" + strings.Join(args, ", ") + ")"
}

// binary operator has slightly different formatting
type BinOp struct { *baseFunc }

func Operator(name string) BinOp {
    return BinOp{ &baseFunc{name, 2} }
}

func (b BinOp) Format(args ...string) string {
    return "(" + args[0] + " " + b.name + " " + args[1] + ")"
}

// input variable type
type Var struct {
    *baseFunc
    Narg int
//    val  *Value
}

func Variable(name string, narg int) Var {
    return Var{ &baseFunc{name,0}, narg } //, &(pset.inputs[narg]) }
}

func (v Var) Eval(input ...Value) Value { return input[v.Narg] }

func (v Var) Format(args ...string) string { return v.name }

// an ephemeral constant holds both a value and a terminal function to update it
type EphemeralConstant struct {
    *baseFunc
    op  Opcode
    val Value
}

func Ephemeral(name string, op Opcode) EphemeralConstant {
    return EphemeralConstant{ &baseFunc{name,0}, op, nil }
}

func (o EphemeralConstant) Eval(args ...Value) Value { return o.val }

func (o EphemeralConstant) Init() EphemeralConstant{
    return EphemeralConstant{ o.baseFunc, o.op, o.op.Eval() }
}

// String shows the name, Format shows the value
func (o EphemeralConstant) Format(args ...string) string {
    return fmt.Sprint(o.val)
}

// create a new primitive set given names of input variables (or empty list if none)
func CreatePrimitiveSet(vars ...string) *PrimSet {
    pset := &PrimSet{}
    pset.primitives = []Opcode{}
    pset.terminals = make([]Opcode, len(vars))
//    pset.inputs = make([]Value, len(vars))
    for i, name := range vars {
        pset.terminals[i] = Variable(name, i)
    }
    return pset
}

// add one or more opcodes to the primitive set
func (pset *PrimSet) Add(ops ...Opcode) {
    for _, op := range ops {
        if op.Arity() > 0 {
            pset.primitives = append(pset.primitives, op)
        } else {
            pset.terminals = append(pset.terminals, op)
        }
    }
}

// returns the nth variable in the primitive set
func (pset *PrimSet) Var(n int) Opcode {
    return pset.terminals[n]
}

// set the nth input value
//func (pset *PrimSet) Set(n int, v Value) {
//    pset.inputs[n] = v
//}

// accessors for Generator interface
func (pset *PrimSet) Terminals() []Opcode {
    return pset.terminals
}

func (pset *PrimSet) Primitives() []Opcode { 
    return pset.primitives
}

// make a copy of expression
func (e Expr) Clone() Expr {
    return append([]Opcode{}, e...)
}

// traverse the tree depth first starting from pos
func (e Expr) Traverse(pos int, nfunc, tfunc func(Opcode)) int {
    op := e[pos]
    arity := op.Arity()
    if arity == 0 {
        if tfunc != nil { tfunc(op) }
    } else {
        for i:=0; i<arity; i++ {
            pos = e.Traverse(pos+1, nfunc, tfunc)
        }
        if nfunc != nil { nfunc(op) }
    }
    return pos
}

// evaluate expression as prefix list
func (e Expr) Eval(input []Value) Value {
    list := []Value{}
    node := func(op Opcode) {
        end := len(list)-op.Arity()
        list = append(list[:end], op.Eval(list[end:]...))
    }
    term := func(op Opcode) {
        list = append(list, op.Eval(input...))
    }
    e.Traverse(0, node, term)
    return list[0]
}

// nicely format the expression
func (e Expr) Format() string {
    list := []string{}
    node := func(op Opcode) {
        end := len(list)-op.Arity()
        list = append(list[:end], op.Format(list[end:]...))
    }
    term := func(op Opcode) {
        list = append(list, op.Format())
    }
    e.Traverse(0, node, term)
    return list[0]
}

// maximum height of tree from root 
func (e Expr) Depth() int {
    stack := make([]int, 1, len(e))
    maxDepth, depth := 0, 0
    stack[0] = 0
    for _, op := range e {
        end := len(stack)-1
        depth, stack = stack[end], stack[:end]
        if depth > maxDepth { maxDepth = depth }
        for i:=0; i<op.Arity(); i++ {
            stack = append(stack, depth+1)
        }
    }
    return maxDepth
}



