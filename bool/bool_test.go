package bool
import (
	"testing"
    "github.com/jnb666/gogp/expr"
    "github.com/jnb666/gogp/gp"
)

// test evaluating some simple expressions
func TestEval(t *testing.T) {
    pset := gp.CreatePrimSet("A", "B")
    pset.Add(True, False, And, Or, Xor, Not)
    exprs := []expr.Expr{
        expr.Expr{ Xor, And, True, False, Or, True, False, True },
        expr.Expr{ And, pset.Var(0), Not, pset.Var(1) },
    }

    val := exprs[0].Eval([]expr.Value{False, False})
    t.Log(exprs[0], exprs[0].Format(), " => ", val)
    if val != True { t.Errorf("Eval(%s) = %v", exprs[0], val) }

    input := [][]expr.Value{ 
        []expr.Value{False, False}, 
        []expr.Value{False, True}, 
        []expr.Value{True, False}, 
        []expr.Value{True, True},
    }
    for i, expect := range []expr.Value{ False, False, True, False } {
        val := exprs[1].Eval(input[i])
        t.Log("input:", input[i], exprs[1].Format(), "->", val)
        if val != expect { t.Errorf("Eval(%s) = %v", exprs[1], val) }
    }
}
