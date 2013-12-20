// Package num provides a floating point numeric type and associated operations.
package num
import (
    "fmt"
    "github.com/jnb666/gogp/gp"
)

const DIVIDE_PROTECT = 1e-10

var (
    Add = Operator("+", func(a, b Num)Num { return a+b })
    Sub = Operator("-", func(a, b Num)Num { return a-b })
    Mul = Operator("*", func(a, b Num)Num { return a*b })
    Div = Operator("/", pdiv)
    Neg = UnaryFunc("-", func(a Num)Num { return -a })
)

// Num is a floating point value which implements the gp.Opcode interface
type Num float64

func (n Num) Arity() int { return 0 }

func (n Num) Eval(args ...gp.Value) gp.Value { return n }

func (n Num) String() string { return fmt.Sprint(float64(n)) }

func (n Num) Format(args ...string) string { return n.String() }

// Term is a numeric function with no arguments
type Term struct{
    gp.Opcode
    fun func() Num
}

func Terminal(name string, fun func()Num) Term {
    return Term{ gp.Terminal(name), fun }
}

func (o Term) Eval(args ...gp.Value) gp.Value { 
    return o.fun()
}

// Unary is a numeric function with one argument
type Unary struct{
    gp.Opcode
    fun func(a Num) Num
}

func UnaryFunc(name string, fun func(a Num)Num) Unary {
    return Unary{ gp.Function(name,1), fun }
}

func (o Unary) Eval(args ...gp.Value) gp.Value { 
    return o.fun(args[0].(Num))
}

// BinOp is a numeric operator with two arguments
type BinOp struct{
    gp.Opcode
    fun func(a, b Num) Num
}

func Operator(name string, fun func(a, b Num)Num) BinOp {
    return BinOp{ gp.Operator(name), fun }
}

func (o BinOp) Eval(args ...gp.Value) gp.Value { 
    return o.fun(args[0].(Num), args[1].(Num))
}

func pdiv(a, b Num) Num {
	if b > -DIVIDE_PROTECT && b < DIVIDE_PROTECT { return 0 }
    return Num(a / b)
}

