// Package bool provides a boolean type and logic operations for gogp.
package bool
import (
    "github.com/jnb666/gogp/expr"
)

var (
    True  = V(true)
    False = V(false)
    And = Op("and", func(a, b V)V { return a && b })
    Or = Op("or", func(a, b V)V { return a || b })
    Xor = Op("xor", func(a, b V)V { return (a || b) && !(a && b) })
    Not = Unary("not", func(a V)V { return !a })
)

// V is a boolean value which implements the expr.Opcode interface
type V bool

// Arity method returns the number of arguments for the opcode
func (b V) Arity() int { return 0 }

// Eval method for boolean opcode returns the boolean value
func (b V) Eval(args ...expr.Value) expr.Value { return b }

// String method returns the name of the opcode
func (b V) String() string { if b { return "T" } else { return "F" } }

// Format method is called by Expr Format() to return a expression in a human readable format
func (b V) Format(args ...string) string { return b.String() }

// Func constructor returns a boolean function with given arity which implements the expr.Opcode interface
func Func(name string, arity int, fun func([]V) V) expr.Opcode {
    return boolFunc{ expr.Function(name,arity), fun }
}

type boolFunc struct{
    expr.Opcode
    fun func([]V) V
}

func (o boolFunc) Eval(iargs ...expr.Value) expr.Value {
    args := make([]V, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(V)
    }
    return o.fun(args)
}

// Term constructor returns a boolean terminal operator which implements the expr.Opcode interface
func Term(name string, fun func() V) expr.Opcode {
    return termOp{ expr.Function(name,0), fun }
}

type termOp struct{
    expr.Opcode
    fun func() V
}

func (o termOp) Eval(args ...expr.Value) expr.Value {
    return o.fun()
}

// Unary constructor returns a boolean unary operator which implements the expr.Opcode interface
func Unary(name string, fun func(a V) V) expr.Opcode {
    return unaryOp{ expr.Function(name,1), fun }
}

type unaryOp struct{
    expr.Opcode
    fun func(a V) V
}

func (o unaryOp) Eval(args ...expr.Value) expr.Value {
    return o.fun(args[0].(V))
}

// Op constructor returns a boolean binary operator which implements the expr.Opcode interface
func Op(name string, fun func(a, b V) V) expr.Opcode {
    return binOp{ expr.Operator(name), fun }
}

type binOp struct{
    expr.Opcode
    fun func(a, b V) V
}

func (o binOp) Eval(args ...expr.Value) expr.Value {
    return o.fun(args[0].(V), args[1].(V))
}


