// Package num provides a floating point numeric type and associated operations for gogp.
package num
import (
    "fmt"
    "github.com/jnb666/gogp/expr"
)

const DIVIDE_PROTECT = 1e-10

var (
    Add = Op("+", func(a []V)V { return a[0] + a[1] })
    Sub = Op("-", func(a []V)V { return a[0] - a[1] })
    Mul = Op("*", func(a []V)V { return a[0] * a[1] })
    Div = Op("/", protected_divide)
    Neg = Func("-", 1, func(a []V)V { return -a[0] })
)

func protected_divide(a []V) V {
	if a[1] > -DIVIDE_PROTECT && a[1] < DIVIDE_PROTECT { return 0 }
    return V(a[0] / a[1])
}

// V is a floating point value which implements the gp.Opcode interface
type V float64

// Arity method returns the number of arguments for the opcode
func (n V) Arity() int { return 0 }

// Eval method for a numeric operator returns the numeric value
func (n V) Eval(args ...expr.Value) expr.Value { return n }

// String method returns the name of the opcode
func (n V) String() string { return fmt.Sprint(float64(n)) }

// Format method is called by Expr Format() to return a expression in a human readable format
func (n V) Format(args ...string) string { return n.String() }

type numFunc struct{
    expr.Opcode
    fun func([]V) V
}

// Func constructor returns a numeric function with given arity 
// which implements the gp.Opcode interface
func Func(name string, arity int, fun func([]V)V) expr.Opcode {
    return numFunc{ expr.Function(name,arity), fun }
}

// Op constructor returns a numeric binary operator which implements the gp.Opcode interface
func Op(name string, fun func([]V)V) expr.Opcode {
    return numFunc{ expr.Operator(name), fun }
}

func (o numFunc) Eval(iargs ...expr.Value) expr.Value {
    args := make([]V, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(V)
    }
    return o.fun(args)
}

