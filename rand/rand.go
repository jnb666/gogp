// Package rand provides random number generation functions for gogp.
package rand
import (
    "fmt"
    "sync"
    "math/rand"
    "math/big"
    crand "crypto/rand"
)

type Rand struct {
    sync.Mutex
    rng *rand.Rand
    src rngSource
}

var gen Rand

// SetSeed sets the random number seed to seed, or to a random value if seed is <= 0
func Seed(seed int64) int64 {
    if seed <= 0 {
        max := big.NewInt(2<<31-1)
        rseed, _ := crand.Int(crand.Reader, max)
        seed = rseed.Int64()
    }
    fmt.Println("set random seed:", seed)
    gen.Lock()
    gen.src.Seed(seed)
    gen.rng = rand.New(&gen.src)
    gen.Unlock()
    return seed
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0)
func Float64() (r float64) {
    gen.Lock()
    r = gen.rng.Float64()
    gen.Unlock()
    return
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n)
func Intn(n int) (r int) {
    gen.Lock()
    r = gen.rng.Intn(n)
    gen.Unlock()
    return
}

// Save the state of the random source
func Save() (s rngSource) {
    gen.Lock()
    s = gen.src
    gen.Unlock()
    return
}

// Restore the state of the random source
func Restore(s rngSource) {
    gen.Lock()
    gen.src = s
    gen.rng = rand.New(&gen.src)
    gen.Unlock()
}
