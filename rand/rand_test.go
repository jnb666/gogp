package rand
import (
    "testing"
)

func TestRand(t *testing.T) {
    Seed(0)
    s := Save()
    vals := make([]float64, 8)
    for i := range vals {
        vals[i] = Float64()
    }
    t.Log(vals)
    Restore(s)
    for i := range vals {
        r := Float64()
        t.Log(r)
        if r != vals[i] { t.Error("should have same vals", r, vals[i]) }
    }
}