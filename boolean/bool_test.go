package boolean

import (
	"github.com/jnb666/gogp/gp"
	"testing"
)

type pair struct{ A, B V }

// test evaluating some simple expressions
func TestEval(t *testing.T) {
	pset := gp.CreatePrimSet(2, "A", "B")
	pset.Add(True, False, And, Or, Xor, Not)
	exprs := []gp.Expr{
		{Xor, And, True, False, Or, True, False, True},
		{And, pset.Var(0), Not, pset.Var(1)},
	}

	val := exprs[0].Eval(False, False)
	t.Log(exprs[0], exprs[0].Format(), " => ", val)
	if val != True {
		t.Errorf("Eval(%s) = %v", exprs[0], val)
	}

	input := []pair{{False, False}, {False, True}, {True, False}, {True, True}}

	for i, expect := range []gp.Value{False, False, True, False} {
		val := exprs[1].Eval(input[i].A, input[i].B)
		t.Log("input:", input[i], exprs[1].Format(), "->", val)
		if val != expect {
			t.Errorf("Eval(%s) = %v", exprs[1], val)
		}
	}
}
