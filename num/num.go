// Package num provides a floating point numeric type and associated operations for gogp.
package num
import (
    "fmt"
    "github.com/jnb666/gogp/expr"
)

const DIVIDE_PROTECT = 1e-10

var (
    Add = Op("+", func(a, b V)V { return a + b })
    Sub = Op("-", func(a, b V)V { return a - b })
    Mul = Op("*", func(a, b V)V { return a * b })
    Div = Op("/", protected_divide)
    Neg = Unary("-", func(a V)V { return -a })
)

func protected_divide(a, b V) V {
	if b > -DIVIDE_PROTECT && b < DIVIDE_PROTECT { return 0 }
    return V(a / b)
}

// V is a floating point value which implements the expr.Opcode interface
type V float64

// Arity method returns the number of arguments for the opcode
func (n V) Arity() int { return 0 }

// Eval method for a numeric operator returns the numeric value
func (n V) Eval(args ...expr.Value) expr.Value { return n }

// String method returns the name of the opcode
func (n V) String() string { return fmt.Sprint(float64(n)) }

// Format method is called by Expr Format() to return a expression in a human readable format
func (n V) Format(args ...string) string { return fmt.Sprint(float64(n)) }

// Ephemeral constructor to create a numeric EphemeralConstant
func Ephemeral(name string, gen func() V) expr.EphemeralConstant {
    return erc{ gen:gen, name:name }
}

type erc struct {
    V
    gen func() V
    name string
}

func (e erc) Init() expr.EphemeralConstant {
    return erc{ e.gen(), e.gen, e.name }
}

func (e erc) String() string {
    return e.name
}

// Func constructor returns a numeric function with given arity 
// which implements the gp.Opcode interface
func Func(name string, arity int, fun func([]V)V) expr.Opcode {
    return numFunc{ expr.Function(name,arity), fun }
}

type numFunc struct{
    expr.Opcode
    fun func([]V) V
}

func (o numFunc) Eval(iargs ...expr.Value) expr.Value {
    args := make([]V, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(V)
    }
    return o.fun(args)
}

// Term constructor returns a numeric terminal operator which implements the expr.Opcode interface
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

// Unary constructor returns a numeric unary operator which implements the expr.Opcode interface
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

// Op constructor returns a numeric binary operator which implements the expr.Opcode interface
func Op(name string, fun func(a, b V)V) expr.Opcode {
    return numOp{ expr.Operator(name), fun }
}

type numOp struct{
    expr.Opcode
    fun func(a, b V) V
}

func (o numOp) Eval(args ...expr.Value) expr.Value {
    return o.fun(args[0].(V), args[1].(V))
}

