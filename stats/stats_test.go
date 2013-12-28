package stats
import (
	"testing"
    "math/rand"
    "fmt"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/num"
)

var exp = `Gen      Evals    FitMax   FitAvg   FitStd   SizeAvg  SizeMax  DepthAvg DepthMax
0        1000     1        0.493    0.285    6.74     15       1.94     3`

var fields = []string{"Gen", "Evals", "Fit.Max", "Fit.Avg", "Fit.Std", "Size.Max", "Depth.Max" }

func getStats(t *testing.T, gen int) *Stats {
    pset := gp.CreatePrimSet(1, "x")
    pset.Add(num.Add, num.Sub, num.Mul, num.Div, num.Neg, num.V(0), num.V(1))
    pop := gp.CreatePopulation(1000, gp.GenFull(pset, 1, 3))
    for i := range pop {
        pop[i].Fitness = rand.Float64()
    }
    s := Create(pop, gen, len(pop))
    t.Log(s)
    return s
}

// test generating population and calculating stats
func TestStats(t *testing.T) {
    gp.SetSeed(1)
    s := getStats(t, 0)
    if fmt.Sprint(s) != exp {
        t.Error("stats text looks wrong! Expected\n", exp)
    }
    for _, fld := range fields {
        val, err := s.Get(fld)
        if err != nil { t.Error(err) }
        t.Log(fld, "=>" , val)
    }
    _, err := s.Get("Fit.Foo")
    if fmt.Sprint(err) != "Stats field Foo is not valid" { 
        t.Error("expected error for missing field")
    }
}


