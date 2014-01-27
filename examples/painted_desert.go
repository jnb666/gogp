package main
// painted desert example
// from Evaluation of Emergent Cooperative Behaviour using Genetic Progamming [Koza 1993]

import (
    "fmt"
    "flag"
    "strings"
    "math/rand"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
    "github.com/jnb666/gogp/util"
)

// colour mappings (-1, 0, 1, 2)
var (
    Colors   = []byte{'.', 'R', 'G', 'B'}
    ColorsLC = []byte{'+', 'r', 'g', 'b'}
)

// grid of cells and other global vars
type Grid struct {
    colors [][]byte
    ants   []*Ant
    ant    *Ant
    rng    *rand.Rand
    steps, maxMoves, id int
    path   [][4]int
}

// per ant data
type Ant struct {
    row, col, carrying, moves int
    pickupRow, pickupCol int
}

// get ant at given location or nil if none
func (g *Grid) at(row, col int) *Ant {
    for _, ant := range g.ants {
        if ant.col == col && ant.row == row {
            return ant
        }
    }
    return nil
}

// print out the grid
func (g *Grid) String() string {
    lines := make([]string, len(g.colors)+1)
    for y, row := range g.colors {
        for x, cell := range row {
            // ground
            lines[y] += " " + string(cell)
            // ant
            if ant := g.at(y,x); ant != nil {
                lines[y] += string(ColorsLC[ant.carrying+1])
            } else {
                lines[y] += " "
            }
        }
    }
    return strings.Join(lines, "\n")
}

// deep copy of grid data
func (g *Grid) clone() *Grid {
    grid := *g
    grid.colors = make([][]byte, len(g.colors))
    for row, line := range g.colors {
        grid.colors[row] = append([]byte{}, line...)
    }
    grid.ants = make([]*Ant, len(g.ants))
    for i, ant := range g.ants {
        grid.ants[i] = new(Ant)
        *(grid.ants[i]) = *ant
    }
    grid.path = make([][4]int, 0, g.maxMoves)
    return &grid
}

// get next position
func (g *Grid) next(row, col, dir int) (nRow, nCol int) {
    rows, cols := len(g.colors), len(g.colors[0])
    nRow = util.Mod(row + []int{1, 0, -1, 0}[dir], rows)
    nCol = util.Mod(col + []int{0, 1, 0, -1}[dir], cols)
    return
}

// get color of cell at given position, 
func (g *Grid) color(row, col int) (val int) {
    switch g.colors[row][col] {
        case 'R': val = 0
        case 'G': val = 1
        case 'B': val = 2
        default: val = -1
    }
    return
}

// read the trail file to setup the grid
func readGrid(file string) *Grid {
    s := util.Open(file)
    grid := Grid{}
    // first line has config params
    var numAnts int
    util.Read(s, &numAnts, &grid.steps, &grid.maxMoves)
    fmt.Printf("numAnts=%d steps=%d maxMoves=%d\n", numAnts, grid.steps, grid.maxMoves)
    // read initial ant positions
    grid.ants = make([]*Ant, numAnts)
    for i := range grid.ants {
        ant := Ant{ carrying:-1 }
        util.Read(s, &ant.col, &ant.row)
        grid.ants[i] = &ant
    }
    // read the grid colors
    grid.colors = [][]byte{}
    for s.Scan() {
        line := append([]byte{}, s.Bytes()...)
        grid.colors = append(grid.colors, line)
    }
    return &grid
}

// terminal set
type terminal struct {
    *gp.BaseFunc
    fn func(*Grid)int
}

func Terminal(name string, fn func(*Grid)int) gp.Opcode {
    return terminal{&gp.BaseFunc{name,0}, fn}
}

func (o terminal) Eval(args ...gp.Value) gp.Value {
    return o.fn
}

// move in given direction
func move(dir int) func(*Grid) int {
    return func(g *Grid) int {
        if g.ant.moves < g.maxMoves {
            g.ant.moves++
            row, col := g.next(g.ant.row, g.ant.col, dir)
            if g.at(row, col) == nil {
                g.ant.row, g.ant.col = row, col
                //if g.ant.carrying >= 0 {
                //    g.path = append(g.path, [4]int{g.id, col, row, 0})
                //}
            }
        }
        return g.color(g.ant.row, g.ant.col)
    }
}

// pick up grain at current pos if not currently carrying anything
func pickUp(g *Grid) int {
    color := g.color(g.ant.row, g.ant.col)
    if g.ant.moves < g.maxMoves && g.ant.carrying < 0 && color >= 0 {
        g.ant.moves++
        g.ant.carrying = color
        g.ant.pickupRow, g.ant.pickupCol = g.ant.row, g.ant.col
        g.colors[g.ant.row][g.ant.col] = '.'
        g.path = append(g.path, [4]int{g.id, g.ant.col, g.ant.row, color+1})
        return -1
    }
    return color
}

// function set
type ifelse struct {
    *gp.BaseFunc
    cond func(*Grid, ...gp.Value) bool
}

func IfElse(name string, arity int, cond func(*Grid, ...gp.Value)bool) gp.Opcode {
    return ifelse{&gp.BaseFunc{name, arity}, cond}
}

func (o ifelse) Eval(arg ...gp.Value) gp.Value {
    arity := o.Arity()
    return func(g *Grid) int {
        if o.cond(g, arg...) {
            return arg[arity-2].(func(*Grid)int)(g)
        } else {
            return arg[arity-1].(func(*Grid)int)(g)
        }
    }
}

// if arg0 <= arg1 then arg2 else arg3
func ifLessThanOrEqual(g *Grid, arg ...gp.Value) bool {
    return arg[0].(func(*Grid)int)(g) <= arg[1].(func(*Grid)int)(g)
}

// if arg0 < 0 then arg1 else arg2
func ifLessThanZero(g *Grid, arg ...gp.Value) bool {
    return arg[0].(func(*Grid)int)(g) < 0
}

// try and drop a grain
func (g *Grid) drop(row, col, color int) bool {
    if color >= 0 && g.color(row, col) < 0 {
        g.colors[row][col] = Colors[color+1]
        g.path = append(g.path, [4]int{g.id, col, row, -(color+1)})
        return true
    }
    return false
}


// if carrying a grain and current position is empty drop and call arg0, else call arg1
func ifDrop(g *Grid, arg ...gp.Value) bool {
    if g.ant.moves < g.maxMoves && g.drop(g.ant.row, g.ant.col, g.ant.carrying) {
        g.ant.moves++
        g.ant.carrying = -1
        return true
    }
    return false
}

// run the code - step each ant in turn
func run(g *Grid, code gp.Expr) *Grid {
    runFunc := code.Eval().(func(*Grid)int)
    grid := g.clone()
    // always use same random number set so we get a consistent fitness score
    grid.rng = rand.New(rand.NewSource(1))
    for i := 0; i < grid.steps; i++ {
        for id := range grid.ants {
            grid.id = id
            grid.ant = grid.ants[id]
            if grid.ant.moves < grid.maxMoves {
                runFunc(grid)
            }
        }
    }
    // if ant is holding a grain it must drop it so that it can be counted
    for id, ant := range grid.ants {
        if ant.carrying < 0 { continue }
        grid.id = id
        // if current location is empty, count it here
        if grid.drop(ant.row, ant.col, ant.carrying) { continue }
        // if cell we picked it up from is empty, count it there
        if grid.drop(ant.pickupRow, ant.pickupCol, ant.carrying) { continue }
        // find next free space
        grid.mustDrop(ant.row, ant.col, ant.carrying)
    }
    // show final location
    for id, ant := range grid.ants {
        grid.path = append(grid.path, [4]int{id, ant.col, ant.row, 0})
    }
    return grid
}

// force drop in next free space
func (g *Grid) mustDrop(row, col, color int) {
    rows, cols := len(g.colors), len(g.colors[0])
    for xoff := 1; xoff < cols; xoff++ {
        x := util.Mod(col+xoff, cols)
        for yoff := 0; yoff <= xoff; yoff++ {
            y := util.Mod(row+yoff, rows)
            if g.drop(y, x, color) { return }
            if yoff > 0 {             
                y := util.Mod(row-yoff, rows)
                if g.drop(y, x, color) { return }
            }
        } 
    }
    panic("nowhere to drop the sand!")
}

// Raw fitness is the product of the color of the grain of sand (1,2,3) and the distance
// between the grain and the Y-axis when execution of the particular program ceases.
func fitnessFunc(g *Grid) func(gp.Expr) (float64, bool) {
    return func(code gp.Expr) (float64, bool) {
        grid := run(g, code)
        fit := 0
        // check every square for sand on the ground
        for row, line := range grid.colors {
            for col := range line {
                switch grid.color(row, col) {
                    case 0: fit += 3*(col+1)
                    case 1: fit += 2*(col+1)
                    case 2: fit += col+1
                }
            }
        }
        // scale fitness to 0 to 1 range
        return 100/float64(fit), true
    }
}

// returns function to plot path of best individual
func createPlot(g *Grid, size, delay int) func(gp.Population) []byte {
    styles := []string{"fill:grey", "fill:red", "fill:green", "fill:blue"}
    rows, cols := len(g.colors), len(g.colors[0])
    cellSize := size/cols
    return func(pop gp.Population) []byte {
        ch := make(chan [][4]int)
        go func() {
            grid := run(g, pop.Best().Code)
            ch <- grid.path
        }()
        // draw grid
        plot := util.SVGPlot(size, size, cellSize)
        plot.AddGrid(cols, rows, delay, func(x, y int) string { 
            return styles[g.color(y,x)+1]
        })
        // draw ants
        for i, ant := range g.ants {
            id := fmt.Sprintf("ant%d", i)
            plot.AddCircle(id, ant.col, ant.row, "fill:none")
        }
        plot.AnimateMulti("ant", <-ch, styles)
        return plot.Data()
    }
}

// build and run model
func main() {
    // get options
    var maxSize, maxDepth int
    var configFile string
    flag.IntVar(&maxSize, "size", 0, "maximum tree size - zero for none")
    flag.IntVar(&maxDepth, "depth", 0, "maximum tree depth - zero for none")
    flag.StringVar(&configFile, "config", "desert.txt", "grid definition file")    
    opts := util.DefaultOptions
    util.ParseFlags(&opts)

    // create primitive set
    grid := readGrid(configFile)
    pset := gp.CreatePrimSet(0)
    pset.Add(Terminal("x", func(g *Grid)int { return g.ant.col }))
    pset.Add(Terminal("y", func(g *Grid)int { return g.ant.row }))
    pset.Add(Terminal("carrying", func(g *Grid)int { return g.ant.carrying }))
    pset.Add(Terminal("color", func(g *Grid)int { return g.color(g.ant.row, g.ant.col) }))
    pset.Add(Terminal("go-n", move(0)))
    pset.Add(Terminal("go-e", move(1)))
    pset.Add(Terminal("go-s", move(2)))
    pset.Add(Terminal("go-w", move(3)))
    pset.Add(Terminal("go-rand", func (g *Grid) int { return move(g.rng.Intn(4))(g) }))
    pset.Add(Terminal("pickup", pickUp))
    pset.Add(IfElse("iflte", 4, ifLessThanOrEqual))
    pset.Add(IfElse("ifltz", 3, ifLessThanZero))
    pset.Add(IfElse("ifdrop", 2, ifDrop))

    // setup model
    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 1, 2),
        PopSize: opts.PopSize,
        Fitness: fitnessFunc(grid),
        Offspring: gp.Tournament(opts.TournSize),
        Mutate: gp.MutUniform(gp.GenFull(pset, 0, 2)),
        MutateProb: opts.MutateProb,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: opts.CrossoverProb,
        Threads: opts.Threads,
    }
    if maxDepth > 0 {
        problem.AddDecorator(gp.DepthLimit(maxDepth))
    }
    if maxSize > 0 {
        problem.AddDecorator(gp.SizeLimit(maxSize))
    }
    problem.PrintParams("== Artificial ant ==")

    logger := stats.NewLogger(opts.MaxGen, opts.TargetFitness)
    if opts.Verbose {
        logger.OnDone = func(best *gp.Individual) {
            g := run(grid, best.Code)
            fmt.Println(g)     
        }
    }

    // run
    if opts.Plot {
        logger.RegisterSVGPlot("best", createPlot(grid, 500, 2))
        stats.MainLoop(problem, logger, ":8080", "../web")
    } else {
        fmt.Println()
        logger.PrintStats = true
        logger.PrintBest = opts.Verbose
        problem.Run(logger)
    }
}