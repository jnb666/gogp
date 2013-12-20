// Package num provides a floating point numeric type and associated operations for gogp.
package num
import (
    "fmt"
    "github.com/jnb666/gogp/gp"
)

const DIVIDE_PROTECT = 1e-10

var (
    Add = NumOp("+", func(a ...Num)Num { return a[0] + a[1] })
    Sub = NumOp("-", func(a ...Num)Num { return a[0] - a[1] })
    Mul = NumOp("*", func(a ...Num)Num { return a[0] * a[1] })
    Div = NumOp("/", pdiv)
    Neg = NumFunc("-", 1, func(a ...Num)Num { return -a[0] })
)

func pdiv(a ...Num) Num {
	if a[1] > -DIVIDE_PROTECT && a[1] < DIVIDE_PROTECT { return 0 }
    return Num(a[0] / a[1])
}

// Num is a floating point value which implements the gp.Opcode interface
type Num float64

// Arity method returns the number of arguments for the opcode
func (n Num) Arity() int { return 0 }

// Eval method for a numeric operator returns the numeric value
func (n Num) Eval(args ...gp.Value) gp.Value { return n }

// String method returns the name of the opcode
func (n Num) String() string { return fmt.Sprint(float64(n)) }

// Format method is called by Expr Format() to return a expression in a human readable format
func (n Num) Format(args ...string) string { return n.String() }

type numFunc struct{
    gp.Opcode
    fun func(...Num) Num
}

// NumFunc constructor returns a numeric function with given arity 
// which implements the gp.Opcode interface
func NumFunc(name string, arity int, fun func(...Num)Num) gp.Opcode {
    return numFunc{ gp.Function(name,arity), fun }
}

// NumOp constructor returns a numeric binary operator which implements the gp.Opcode interface
func NumOp(name string, fun func(...Num)Num) gp.Opcode {
    return numFunc{ gp.Operator(name), fun }
}

func (o numFunc) Eval(iargs ...gp.Value) gp.Value {
    args := make([]Num, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(Num)
    }
    return o.fun(args...)
}




