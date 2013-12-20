// Package bool provides a boolean type and logic operations for gogp.
package bool
import (
    "fmt"
    "github.com/jnb666/gogp/gp"
)

var (
    True  = Bool(true)
    False = Bool(false)
    And = BoolOp("and", func(a, b Bool)Bool { return a && b })
    Or  = BoolOp("or", func(a, b Bool)Bool { return a || b })
    Not = BoolFunc("not", 1, func(a Bool)Bool { return !a })
)

// Bool is a boolean value which implements the gp.Opcode interface
type Bool bool

// Arity method returns the number of arguments for the opcode
func (b Bool) Arity() int { return 0 }

// Eval method for boolean opcode returns the boolean value
func (b Bool) Eval(args ...gp.Value) gp.Value { return b }

// String method returns the name of the opcode
func (b Bool) String() string { return fmt.Sprint(bool(b)) }

// Format method is called by Expr Format() to return a expression in a human readable format
func (b Bool) Format(args ...string) string { return b.String() }

type boolFunc struct{
    gp.Opcode
    fun func(...Bool) Bool
}

// BoolFunc constructor returns a boolean function with given arity which implements 
// the gp.Opcode interface
func BoolFunc(name string, arity int, fun func(...Bool)Bool) gp.Opcode {
    return boolFunc{ gp.Function(name, arity), fun }
}

// BoolOp constructor returns a boolean binary operator which implements the gp.Opcode interface
func BoolOp(name string, fun func(...Bool)Bool) gp.Opcode {
    return boolFunc{ gp.Operator(name), fun }
}

func (o boolFunc) Eval(iargs ...gp.Value) gp.Value {
    args := make([]Bool, len(iargs))
    for i, iarg := range iargs {
        args[i] = iarg.(Bool)
    }
    return o.fun(args...)
}


