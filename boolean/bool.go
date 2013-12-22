// Package boolean provides a boolean type and logic operations for gogp.
package boolean
import (
    "github.com/jnb666/gogp/gp"
)

var (
    True  = V(true)
    False = V(false)
    And = Op("and", func(a, b V)V { return a && b })
    Or = Op("or", func(a, b V)V { return a || b })
    Xor = Op("xor", func(a, b V)V { return (a || b) && !(a && b) })
    Not = Unary("not", func(a V)V { return !a })
)

// V is a boolean value which implements the gp.Opcode interface
type V bool

// Arity method returns the number of arguments for the opcode
func (b V) Arity() int { return 0 }

// Eval method for boolean opcode returns the boolean value
func (b V) Eval(args ...gp.Value) gp.Value { return b }

// String method returns the name of the opcode
func (b V) String() string { if b { return "true" } else { return "false" } }

// Format method is called by Expr Format() to return a expression in a human readable format
func (b V) Format(args ...string) string { return b.String() }

// Func constructor returns a boolean function with given arity which implements the gp.Opcode interface
func Func(name string, arity int, fun func([]V) V) gp.Opcode {
    return boolFunc{ gp.Function(name,arity), fun }
}

type boolFunc struct{
    gp.Opcode
    fun func([]V) V
}

func (o boolFunc) Eval(iargs ...gp.Value) gp.Value {
    args := make([]V, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(V)
    }
    return o.fun(args)
}

// Term constructor returns a boolean terminal operator which implements the gp.Opcode interface
func Term(name string, fun func() V) gp.Opcode {
    return termOp{ gp.Terminal(name), fun }
}

type termOp struct{
    gp.Opcode
    fun func() V
}

func (o termOp) Eval(args ...gp.Value) gp.Value {
    return o.fun()
}

// Unary constructor returns a boolean unary operator which implements the gp.Opcode interface
func Unary(name string, fun func(a V) V) gp.Opcode {
    return unaryOp{ gp.Function(name,1), fun }
}

type unaryOp struct{
    gp.Opcode
    fun func(a V) V
}

func (o unaryOp) Eval(args ...gp.Value) gp.Value {
    return o.fun(args[0].(V))
}

// Op constructor returns a boolean binary operator which implements the gp.Opcode interface
func Op(name string, fun func(a, b V) V) gp.Opcode {
    return binOp{ gp.Operator(name), fun }
}

type binOp struct{
    gp.Opcode
    fun func(a, b V) V
}

func (o binOp) Eval(args ...gp.Value) gp.Value {
    return o.fun(args[0].(V), args[1].(V))
}


