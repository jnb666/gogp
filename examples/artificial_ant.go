package main
// artifical ant example

import (
    "os"
    "bufio"
    "fmt"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
)

const (
    MAX_MOVES = 600
    ROWS = 32
    COLS = 32
    FOOD = '#'
    TRAIL = '*'
    START = 'S'
)

var (
    DIR_ROW = []int{ 1, 0, -1, 0 }
    DIR_COL = []int{ 0, 1, 0, -1 }
)

// positive modulus
func mod(a, b int) int {
    return (a % b + b) % b
}

// grid of cells
type Grid [][]byte

func (g Grid) String() string {
    text := ""
    for _, line := range g {
        text += string(line) + "\n"
    }
    return text
}

// ant data
type Ant struct {
    row, col, dir int
    moves, eaten, totalFood int
    grid Grid
}

// execute each of args in sequence
type progN struct { *gp.BaseFunc }

func (o progN) Eval(args ...gp.Value) gp.Value {
    return func() {
        for _, arg := range args {
            arg.(func())()
        }
    }
}

// terminal node
type terminal struct {
    *gp.BaseFunc
    fn func(*Ant)
    ant *Ant
}

func Terminal(name string, fn func(*Ant), ant *Ant) gp.Opcode {
    return terminal{&gp.BaseFunc{name,0}, fn, ant}
}

func (o terminal) Eval(args ...gp.Value) gp.Value {
    return func() { o.fn(o.ant) }
}

// turn left
func left(ant *Ant) {
    if ant.moves < MAX_MOVES {
        ant.moves++
        ant.dir = mod(ant.dir-1, 4)
        ant.grid[ant.row][ant.col] = TRAIL
    }
}

// turn right
func right(ant *Ant) {
    if ant.moves < MAX_MOVES {
        ant.moves++
        ant.dir = mod(ant.dir+1, 4)
        ant.grid[ant.row][ant.col] = TRAIL
    }
}

// step forward and pick up food if in new cell
func step(ant *Ant) {
    if ant.moves < MAX_MOVES {
        ant.moves++
        ant.row = mod((ant.row + DIR_ROW[ant.dir]), ROWS)
        ant.col = mod((ant.col + DIR_COL[ant.dir]), COLS)
        if ant.grid[ant.row][ant.col] == FOOD { ant.eaten++ }
        ant.grid[ant.row][ant.col] = TRAIL
    }
}

type ifFood struct { 
    *gp.BaseFunc
    ant *Ant
}

// if next cell contains food execute first arg else execute second
func (o ifFood) Eval(args ...gp.Value) gp.Value {
    return func() {
        row := mod(o.ant.row + DIR_ROW[o.ant.dir], ROWS)
        col := mod(o.ant.col + DIR_COL[o.ant.dir], COLS)
        if o.ant.grid[row][col] == FOOD {
            args[0].(func())()
        } else {
            args[1].(func())()
        }
    }
}

// read the trail file to setup the grid
func readTrail(file string) Grid {
    f, err := os.Open(file)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    grid := Grid{}
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        grid = append(grid, scanner.Bytes())
    }
    f.Close()
    return grid
}

// reset ant data to default - make deep copy of grid
func (ant *Ant) reset(grid Grid) {
    ant.grid = make(Grid, len(grid))
    ant.moves = 0
    ant.eaten = 0
    ant.totalFood = 0
    for row, line := range grid {
        ant.grid[row] = append([]byte{}, line...)
        for col, cell := range line {
            switch cell {
            case FOOD:
                ant.totalFood++
            case START:
                ant.row, ant.col = row, col
                ant.dir = 1
            }
        }
    }
}

// run the program to calculate the fitness as no. of food cells eaten / total
func fitnessFunc(ant *Ant, grid Grid) func(gp.Expr) (float64, bool) {
    return func(code gp.Expr) (float64, bool) {
        ant.reset(grid)
        run := code.Eval().(func())
        for ant.moves < MAX_MOVES { run() }
        return float64(ant.eaten) / float64(ant.totalFood), true
    }
}

// build and run model
func main() {
    gp.SetSeed(0)
    grid := readTrail("santafe_trail.txt")
    ant  := &Ant{}

    pset := gp.CreatePrimSet(0)
    pset.Add(progN{ &gp.BaseFunc{"prog2", 2} })
    pset.Add(progN{ &gp.BaseFunc{"prog3", 3} })
    pset.Add(Terminal("left", left, ant))
    pset.Add(Terminal("right", right, ant))
    pset.Add(Terminal("step", step, ant))
    pset.Add(ifFood{ &gp.BaseFunc{"if_food", 2}, ant })

    evalFitness := fitnessFunc(ant, grid)

    problem := gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 1, 2),
        PopSize: 300,
        Fitness: evalFitness,
        Offspring: gp.Tournament(7),
        Mutate: gp.MutUniform(gp.GenFull(pset, 0, 2)),
        MutateProb: 0.2,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: 0.5,
        Threads: 1,
    }

    logger := &stats.Logger{ MaxGen: 40, TargetFitness: 0.99, PrintStats: true }
    pop := problem.Run(logger)
    best := pop.Best()
    fmt.Println(best)
    evalFitness(best.Code)
    fmt.Println(ant.grid)
}



