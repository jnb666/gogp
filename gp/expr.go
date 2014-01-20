package gp
import (
    "strings"
    "github.com/jnb666/gogp/rand"
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

// An EphemeralConstant is a subtype of Opcode. It is typically used to hold a random 
// constant value which is set at when an individual is generated. Implementations of the
// gp.Generator interface should call Init() to get the constant on creation of a new
// individual.
type EphemeralConstant interface {
    Opcode
    Init() EphemeralConstant
}

// The Expr type is defined a slice of Opcodes. An expression is stored internally as a list in prefix
// notation to represent the opcode tree. Methods are provided to evaluate an expression given specified
// terminal nodes.
type Expr []Opcode

// BaseFunc is an abstract implementation of the Opcode interface for embedding in concrete types
type BaseFunc struct {
    OpName  string
    OpArity int
}

func (f *BaseFunc) Arity() int { return f.OpArity }

func (f *BaseFunc) String() string { return f.OpName }

func (f *BaseFunc) Eval(args ...Value) Value { panic("abstract method!") }

func (f *BaseFunc) Format(args ...string) string {
    if len(args) > 0 {
        return f.OpName + "(" + strings.Join(args, ", ") + ")"
    } else {
        return f.OpName
    }
}

// binary operator type 
type binOp struct { *BaseFunc }

func (b binOp) Format(args ...string) string {
    return "(" + args[0] + " " + b.OpName + " " + args[1] + ")"
}

// variable type
type variable struct {
    *BaseFunc
    Narg int
}

func (v variable) Eval(input ...Value) Value { return input[v.Narg] }

// Terminal constructor. Returns an Opcode representing a function which does not take any arguments.
func Terminal(name string) Opcode {
    return &BaseFunc{name, 0}
}

// Function constructor. Returns an Opcode representing a function with given name and arity.
func Function(name string, arity int) Opcode {
    return &BaseFunc{name, arity}
}

// Operator constructor. Returns an opcode represening a function with two arguments.
// This is formatted in infix notation by the Format method.
func Operator(name string) Opcode {
    return binOp{ &BaseFunc{name, 2} }
}

// Variable constructor. Returns an opcode representing input variable number narg.
func Variable(name string, narg int) Opcode {
    return variable{ &BaseFunc{name, 0}, narg }
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
func (e Expr) Eval(input ...Value) Value {
    var doEval func() Value
    pos := -1
    doEval = func() Value {
        pos++
        op := e[pos]
        arity := op.Arity()
        switch arity {
        case 0:
            return op.Eval(input...)
        case 1:
            return op.Eval(doEval())
        case 2:
            return op.Eval(doEval(), doEval())
        default:
            args := make([]Value, arity)
            for i := range args {
                args[i] = doEval()
            }
            return op.Eval(args...)
        }
    }
    return doEval()
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

// ReplaceSubtree replaces the code at pos with subtree without updating the subtree argument.
func (e Expr) ReplaceSubtree(pos int, subtree Expr) Expr {
    end := e.Traverse(pos, nil, nil)
    tail := subtree
    if end < len(e)-1 {
        tail = append(tail.Clone(), e[end+1:]...)
    }
    return append(e[:pos], tail...)
}

// RandomSubtree returns postion and a copy of nodes in randomly selected subtree of code
func (e Expr) RandomSubtree() (pos int, subtree Expr) {
    pos = rand.Intn(len(e))
    end := e.Traverse(pos, nil, nil)
    subtree = e[pos:end+1].Clone()
    return
}



