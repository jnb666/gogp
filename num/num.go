// floating point numeric types
package num
import (
    "fmt"
    "gogp/gp"
)

const DIVIDE_PROTECT = 1e-10

// floating point value type which implements Opcode interface
type Num float64

func (n Num) Arity() int { return 0 }

func (n Num) Eval(args ...gp.Value) gp.Value { return n }

func (n Num) String() string { return fmt.Sprint(float64(n)) }

func (n Num) Format(args ...string) string { return n.String() }

// terminal functions
type Term struct{
    gp.Term
    fun func() Num
}

func Terminal(name string, fun func()Num) Term {
    return Term{ gp.Terminal(name), fun }
}

func (o Term) Eval(args ...gp.Value) gp.Value { 
    return o.fun()
}

// unary functions
type Unary struct{
    gp.Func
    fun func(a Num) Num
}

func UnaryFunc(name string, fun func(a Num)Num) Unary {
    return Unary{ gp.Function(name,1), fun }
}

func (o Unary) Eval(args ...gp.Value) gp.Value { 
    return o.fun(args[0].(Num))
}

var Neg = UnaryFunc("-", func(a Num)Num { return -a })

// binary operators
type BinOp struct{
    gp.BinOp
    fun func(a, b Num) Num
}

func Operator(name string, fun func(a, b Num)Num) BinOp {
    return BinOp{ gp.Operator(name), fun }
}

func (o BinOp) Eval(args ...gp.Value) gp.Value { 
    return o.fun(args[0].(Num), args[1].(Num))
}

var Add = Operator("+", func(a, b Num)Num { return a+b })
var Sub = Operator("-", func(a, b Num)Num { return a-b })
var Mul = Operator("*", func(a, b Num)Num { return a*b })
var Div = Operator("/", pdiv)

// protected division
func pdiv(a, b Num) Num {
	if b > -DIVIDE_PROTECT && b < DIVIDE_PROTECT { return 0 }
    return Num(a / b)
}

