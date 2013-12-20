package gp
import (
    "strings"
    "fmt"
)

// The Value type is defined as an empty interface hence any type can be used for a specific model.
// Note that typed GP is not yet supported, it's assumed all values will be of the same type 
// for a given data set. See gogp/num for an example of implementing a floating point numeric type.
type Value interface{}

// The basic atom of the model is the Opcode interface. The implemention must supply the Eval 
// method to evalute an opcode, an Arity method to define the number of arguments a String method
// which returns the opcode name and a Format method which returns the opcode in the context of an
// expression in a human readabile format. 
type Opcode interface {
    Arity() int
    Eval(...Value) Value
    String() string
    Format(...string) string
}

// The Expr type is defined a slice of Opcodes. An expression is stored internally as a list in prefix
// notation to represent the opcode tree. Methods are provided to evaluate an expression given specified
// terminal nodes.
type Expr []Opcode

// A PrimSet represents the set of all of primitive Opcodes for a given run. 
type PrimSet struct {
    terminals  []Opcode
    primitives []Opcode
}

type baseFunc struct {
    name  string
    arity int
}

// Terminal constructor. Returns an Opcode representing a function which does not take any arguments.
func Terminal(name string) Opcode {
    return &baseFunc{name, 0}
}

// Function constructor. Returns an Opcode representing a function with given name and arity.
func Function(name string, arity int) Opcode {
    return &baseFunc{name, arity}
}

// Operator constructor. Returns an opcode represening a function with two arguments.
// This is formatted in infix notation by the Format method.
func Operator(name string) Opcode {
    return binOp{ &baseFunc{name, 2} }
}

// Variable constructor. Returns an opcode representing input variable number narg.
func Variable(name string, narg int) Opcode {
    return variable{ &baseFunc{name,0}, narg }
}

// Ephemeral constructor. An Ephemeral holds a "constant" value which is generated when
// the Init method is called on the provided Opcode. Prior to this the value is nil.
func Ephemeral(name string, op Opcode) Opcode {
    return ephemeral{ &baseFunc{name,0}, op, nil }
}

// Arity method returns the number of arguments for the opcode
func (f *baseFunc) Arity() int { return f.arity }

// String method returns the name of the opcode
func (f *baseFunc) String() string { return f.name }

// Eval abstract base method - must be overridden
func (f *baseFunc) Eval(args ...Value) Value { panic("abstract method!") }

// Format method is called by Expr.Format to return a expression in a human readable format
func (f *baseFunc) Format(args ...string) string {
    return f.name + "(" + strings.Join(args, ", ") + ")"
}

type binOp struct { *baseFunc }

func (b binOp) Format(args ...string) string {
    return "(" + args[0] + " " + b.name + " " + args[1] + ")"
}

type variable struct {
    *baseFunc
    Narg int
}

// Eval method for variable returns the associated input value
func (v variable) Eval(input ...Value) Value { return input[v.Narg] }

// Format method is called by Expr.Format to return a expression in a human readable format
func (v variable) Format(args ...string) string { return v.name }

type ephemeral struct {
    *baseFunc
    op  Opcode
    val Value
}

// Eval method for ephemeral returns the value which was generated when Init was called
func (o ephemeral) Eval(args ...Value) Value { return o.val }

// Init method for ephemeral returns a new ephemeral constant with the value set
func (o ephemeral) Init() ephemeral{
    return ephemeral{ o.baseFunc, o.op, o.op.Eval() }
}

// Format method is called by Expr.Format to return a expression in a human readable format
func (o ephemeral) Format(args ...string) string {
    return fmt.Sprint(o.val)
}

// CreatePrimitiveSet constructor takes a list of variable names.
func CreatePrimitiveSet(vars ...string) *PrimSet {
    pset := &PrimSet{}
    pset.primitives = []Opcode{}
    pset.terminals = make([]Opcode, len(vars))
    for i, name := range vars {
        pset.terminals[i] = Variable(name, i)
    }
    return pset
}

// Add method adds a new Opcode to the primitive set.
func (pset *PrimSet) Add(ops ...Opcode) {
    for _, op := range ops {
        if op.Arity() > 0 {
            pset.primitives = append(pset.primitives, op)
        } else {
            pset.terminals = append(pset.terminals, op)
        }
    }
}

// Var returns the nth variable in the primitive set.
func (pset *PrimSet) Var(n int) Opcode {
    return pset.terminals[n]
}

// Terminals returns the list of terminal arity 0 opcodes. 
func (pset *PrimSet) Terminals() []Opcode {
    return pset.terminals
}

// Primitives returns the list of non-terminal opcodes (arity > 0).
func (pset *PrimSet) Primitives() []Opcode { 
    return pset.primitives
}

// Clone makes a copy of an expression.
func (e Expr) Clone() Expr {
    return append([]Opcode{}, e...)
}

// Traverse walks the expression tree using a depth first traversal starting at element pos.
// If not nil then tfunc is called for each leaf node and nfunc is called for each node 
// having one or more child nodes.
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

// Eval evaluates an expression for given input values by calling the Eval method on each Opcode.
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

// Format returns a string representation of an expression.
// It calls the Format method on each Opcode to return a result in infix notation. 
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

// Depth returns the maximum height of the code tree from the root. 
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



