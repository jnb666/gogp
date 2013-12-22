package num
import (
	"testing"
    "math/rand"
    "math"
    "github.com/jnb666/gogp/gp"
)

var (
    Rand = Term("rand", func()V { return V(rand.Float64()) })
    Sqr  = Unary("sqr", func(a V)V { return a*a })
    Floor= Unary("floor", func(a V)V { return V(math.Floor(float64(a))) })
)

// setup primitive set
func initPset(all bool) *gp.PrimSet {
    pset := gp.CreatePrimSet(2, "x", "y")
    pset.Add( Add, Sub, Mul, Div, Neg)
    if all { pset.Add(V(42), Sqr, Rand, Floor) }
    return pset
}

// test expressions
func testExprs(pset *gp.PrimSet) []gp.Expr {
    x, y := pset.Var(0), pset.Var(1)
    exprs := []gp.Expr{
        gp.Expr{ Add, V(9), V(4) },
        gp.Expr{ Add, Sqr, x, Sqr, y },
        gp.Expr{ Div, Mul, Add, V(3), V(4), Sub, V(9), V(6), V(2) },
        gp.Expr{ Div, Add, V(1), Neg, V(1), V(0) },
        gp.Expr{ Floor, Mul, V(10), Rand },
    }
    return exprs
}

// test cloning expression
func TestClone(t *testing.T) {
    pset := initPset(true)
    exprs := testExprs(pset)
    orig  := exprs[2].Format()
    t.Log(exprs[2], orig)
    expr2 := exprs[2].Clone()
    clone := expr2.Format()
    t.Log(expr2, clone)
    if orig != clone { t.Errorf("clone got %s - expected %s", clone, orig) } 
}

// test depth and size attributes
func TestDepth(t *testing.T) {
    pset := initPset(true)
    exprs := testExprs(pset)
    ind := gp.Create(exprs[2])
    depth, size := ind.Depth(), ind.Size()
    t.Logf("%s depth=%d size=%d\n", ind, depth, size)
    if depth!=3 || size!=9 {
        t.Errorf("got depth=%d size=%d - expected 3 9", depth, size) 
    } 
}

// test evaluating some simple expressions
func TestEval(t *testing.T) {
    pset := initPset(true)
    exprs := testExprs(pset)
    expect := []V{13, 25, 10.5, 0, 6}
    rand.Seed(1)
    for i, expect := range expect {
        val := exprs[i].Eval(V(3), V(4))
        t.Log(exprs[i], "(3,4) ", exprs[i].Format(), " => ", val)
    	if val != expect { t.Errorf("Eval(%s) = %f", exprs[i], val) }
    }
}

// test generating random individuals
func TestGenerate(t *testing.T) {
    pset := initPset(false)
    pset.Add(V(0), V(1))
    gen := gp.GenRamped(pset, 1, 3)
    gp.SetSeed(0)
    for i:=0; i<10; i++ {
        ind := gen.Generate()
        res := ind.Code.Eval(V(6), V(7))
        t.Log(ind.Code, ind.Code.Format(), "(6,7) =>", res)
    }
}

// test ephemeral random constants
func TestEphemeral(t *testing.T) {
    pset := initPset(false)
    erc := Ephemeral("ERC", func()V { return V(rand.Intn(10)) })
    pset.Add(erc, erc, erc)
    gen := gp.GenFull(pset, 1, 3)
    gp.SetSeed(2)
    ind := gen.Generate()
    t.Log(ind.Code, ind.Code.Format())
    val := ind.Code.Eval(V(6), V(7))
    t.Log("evals to", val, "for x=6 y=7")
    if val != V(16) { t.Errorf("Eval(%s) = %f", ind.Code, val) }
}

// test mutation
type genProxy struct { expr gp.Expr }

func (g genProxy) String() string { return "genProxy" }

func (g genProxy) Generate() *gp.Individual {
    return &gp.Individual{Code: g.expr}
}

func TestMutate(t *testing.T) {
    // all possible mutations
    mutset := map[string]bool{
        "(9 + 4)": false,
        "((9 + 4) + sqr(y))": false,
        "(sqr((9 + 4)) + sqr(y))": false,
        "(sqr(x) + (9 + 4))": false,
        "(sqr(x) + sqr((9 + 4)))": false,
    }
    pset := initPset(true)
    exprs := testExprs(pset)
    before := gp.Individual{ Code: exprs[1] }
    add := exprs[0]
    gen := genProxy{ add }
    t.Log("mutate: ", before.Code ,"plus", gen.Generate().Code)
    mut := gp.MutUniform(gen)
    rand.Seed(1)
    for i:=0; i<10; i++ {
        after := mut.Variate(gp.Population{before.Clone()})
        t.Log("becomes:", after[0])
        text := after[0].Code.Format()
        if _, ok := mutset[text]; ok {
            mutset[text] = true
        } else {
            t.Error("unexpected mutation", text)
        }
    }
    for key, ok := range mutset {
        if !ok {
            t.Error("missing mutation", key)
        }
    }
}

// test crossover between two trees
func TestCrossover(t *testing.T) {
    pset := initPset(true)
    exprs := testExprs(pset)
    breed := gp.CxOnePoint()
    parent := gp.Population{ gp.Create(exprs[1]), gp.Create(exprs[2]) }
    t.Log("Before:\n", parent[0], "\n", parent[1])
    for i:=0; i<5; i++ {
        child := breed.Variate(parent)
        t.Log("After:", i, "\n", child[0], "\n", child[1])
    }
}


