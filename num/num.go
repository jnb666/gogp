// Package num provides a floating point numeric type and associated operations for gogp.
package num
import (
    "fmt"
    "github.com/jnb666/gogp/gp"
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

// V is a floating point value which implements the gp.Opcode interface
type V float64

// Arity method returns the number of arguments for the opcode
func (n V) Arity() int { return 0 }

// Eval method for a numeric operator returns the numeric value
func (n V) Eval(args ...gp.Value) gp.Value { return n }

// String method returns the name of the opcode
func (n V) String() string { return fmt.Sprint(float64(n)) }

// Format method is called by Expr Format() to return a expression in a human readable format
func (n V) Format(args ...string) string { return fmt.Sprint(float64(n)) }

// Ephemeral constructor to create a numeric EphemeralConstant
func Ephemeral(name string, gen func() V) gp.EphemeralConstant {
    return erc{ gen:gen, name:name }
}

type erc struct {
    V
    gen func() V
    name string
}

func (e erc) Init() gp.EphemeralConstant {
    return erc{ e.gen(), e.gen, e.name }
}

func (e erc) String() string {
    return e.name
}

// Func constructor returns a numeric function with given arity 
// which implements the gp.Opcode interface
func Func(name string, arity int, fun func([]V)V) gp.Opcode {
    return numFunc{ gp.Function(name,arity), fun }
}

type numFunc struct{
    gp.Opcode
    fun func([]V) V
}

func (o numFunc) Eval(iargs ...gp.Value) gp.Value {
    args := make([]V, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(V)
    }
    return o.fun(args)
}

// Term constructor returns a numeric terminal operator which implements the gp.Opcode interface
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

// Unary constructor returns a numeric unary operator which implements the gp.Opcode interface
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

// Op constructor returns a numeric binary operator which implements the gp.Opcode interface
func Op(name string, fun func(a, b V)V) gp.Opcode {
    return numOp{ gp.Operator(name), fun }
}

type numOp struct{
    gp.Opcode
    fun func(a, b V) V
}

func (o numOp) Eval(args ...gp.Value) gp.Value {
    return o.fun(args[0].(V), args[1].(V))
}

