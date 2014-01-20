package main
// painted desert example
// from Evaluation of Emergent Cooperative Behaviour using Genetic Progamming [Koza 1993]

import (
    "fmt"
    "flag"
    "math/rand"
    "strings"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
    "github.com/jnb666/gogp/util"
)

// colour mappings (-1, 0, 1, 2)
var (
    Colors   = []byte{'.', 'R', 'G', 'B'}
    ColorsLC = []byte{'+', 'r', 'g', 'b'}
)

// grid of cells
type Grid struct {
    colors [][]byte
    ants   []*Ant
    steps, maxMoves int
}

// ant data
type Ant struct {
    id, row, col, carrying, moves int
    pickupRow, pickupCol int
}

// print out ant info
func (ant Ant) String() string {
    text := fmt.Sprintf("ant %d: at %d,%d", ant.id, ant.col, ant.row)
    if ant.carrying >= 0 {
        text += fmt.Sprintf(" carrying %d", ant.carrying)
    }
    return text
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
    for i, ptr := range g.ants {
        ant := *ptr
        grid.ants[i] = &ant
    }
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
        ant := Ant{ id:i, carrying:-1 }
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
    fn func(*Grid,int)int
}

func Terminal(name string, fn func(*Grid,int)int) gp.Opcode {
    return terminal{&gp.BaseFunc{name,0}, fn}
}

func (o terminal) Eval(args ...gp.Value) gp.Value {
    return o.fn
}

// current location of ant
var X = Terminal("x", func(g *Grid, id int)int { return g.ants[id].col })
var Y = Terminal("y", func(g *Grid, id int)int { return g.ants[id].row })

// color of grain we are carrying
var CARRYING = Terminal("carrying", func(g *Grid, id int)int { return g.ants[id].carrying })

// color of grain on current square
var COLOR = Terminal("color", func(g *Grid, id int)int { return g.color(g.ants[id].row, g.ants[id].col) })

// move in given direction
func move(dir int) func(*Grid,int)int {
    return func(g *Grid, id int) int {
        ant := g.ants[id]
        if ant.moves < g.maxMoves {
            ant.moves++
            row, col := g.next(ant.row, ant.col, dir)
            if g.at(row, col) == nil {
                ant.row, ant.col = row, col
            }
        }
        return g.color(ant.row, ant.col)
    }
}
var GO_N = Terminal("go-n", move(0))
var GO_E = Terminal("go-e", move(1))
var GO_S = Terminal("go-s", move(2))
var GO_W  = Terminal("go-w", move(3))
var GO_RAND = Terminal("go-rand", func (g *Grid, id int) int { return move(rand.Intn(4))(g,id) })

// pick up grain at current pos if not currently carrying anything
func pickUp(g *Grid, id int) int {
    ant := g.ants[id]
    color := g.color(ant.row, ant.col)
    if ant.moves < g.maxMoves && ant.carrying < 0 && color >= 0 {
        ant.moves++
        // fmt.Printf("%d: picks up %d @ %d,%d\n", ant.id, color, ant.col, ant.row)
        ant.carrying = color
        ant.pickupRow, ant.pickupCol = ant.row, ant.col
        g.colors[ant.row][ant.col] = '.'
        return -1
    }
    return color
}
var PICKUP = Terminal("pickup", pickUp)

// function set
// if arg0 <= arg1 then arg2 else arg3
type iflte struct { *gp.BaseFunc }

func (o iflte) Eval(arg ...gp.Value) gp.Value {
    return func(g *Grid, id int) int {
        if arg[0].(func(*Grid,int)int)(g,id) <= arg[1].(func(*Grid,int)int)(g,id) {
            return arg[2].(func(*Grid,int)int)(g,id)
        } else {
            return arg[3].(func(*Grid,int)int)(g,id)
        }
    }
}
var IFLTE = iflte{ &gp.BaseFunc{"iflte", 4} }

// if arg0 < 0 then arg1 else arg2
type ifltz struct { *gp.BaseFunc }

func (o ifltz) Eval(arg ...gp.Value) gp.Value {
    return func(g *Grid, id int) int {
        if arg[0].(func(*Grid,int)int)(g,id) < 0 {
            return arg[1].(func(*Grid,int)int)(g,id)
        } else {
            return arg[2].(func(*Grid,int)int)(g,id)
        }
    }
}
var IFLTZ = ifltz{ &gp.BaseFunc{"ifltz", 3} }

// if carrying a grain and current position is empty drop and call arg0, else call arg1
type ifdrop struct { *gp.BaseFunc }

func (o ifdrop) Eval(arg ...gp.Value) gp.Value {
    return func(g *Grid, id int) int {
        ant := g.ants[id]
        if ant.moves < g.maxMoves && ant.carrying >= 0 && g.color(ant.row, ant.col) < 0 {
            // fmt.Printf("%d: drops %d @ %d,%d\n", ant.id, ant.carrying, ant.col, ant.row)
            ant.moves++
            g.colors[ant.row][ant.col] = Colors[ant.carrying+1]
            ant.carrying = -1
            return arg[0].(func(*Grid,int)int)(g,id)
        } else {
            return arg[1].(func(*Grid,int)int)(g,id)
        }
    }
}
var IFDROP = ifdrop{ &gp.BaseFunc{"ifdrop", 2} }

// run the code - step each ant in turn
func run(g *Grid, code gp.Expr) *Grid {
    runFunc := code.Eval().(func(*Grid,int)int)
    grid := g.clone()
    fmt.Println("*BEFORE*", code)
    fmt.Println(grid)
    for i := 0; i < grid.steps; i++ {
        for j, ant := range grid.ants {
            if ant.moves < grid.maxMoves {
                runFunc(grid, j)
            }
        }
    }
    fmt.Println("*AFTER*")
    fmt.Println(grid)
    return grid
}

// convert color and column to fitness
func fitnessVal(color int, xpos int) (fit int) {
    switch color {
        case 0: fit = xpos+1
        case 1: fit = 2*(xpos+1)
        case 2: fit = 3*(xpos+1)
    }
    return
}

// get column for sand ant is carrying for fitbess score
func getColumn(ant *Ant, grid *Grid) int {
    if grid.color(ant.row, ant.col) < 0 {
        // current location is empty, count it here
        return ant.col
    }
    if grid.color(ant.pickupRow, ant.pickupCol) < 0 {
        // cell we picked it up from is empty, count it there
        return ant.pickupCol
    }
    // find next free space
    for col := ant.col+1; col < len(grid.colors[0]); col++ {
        if grid.color(ant.row, col) < 0 {
            return col
        }
    }
    return len(grid.colors[0])
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
                color := grid.color(row, col)
                if color >= 0 {
                    fit += fitnessVal(color, col)
                }
            }
        }
        // check ants and count any sand they're carrying
        for _, ant := range grid.ants {
            if ant.carrying >= 0 {
                fit += fitnessVal(ant.carrying, getColumn(ant, grid))
            }
        }
        // scale fitness to 0 to 1 range
        normFit := 100/float64(fit)
        fmt.Println(code.Format(), fit)
        return normFit, true
    }
}

// function to plot grid colors
func plotGrid(g *Grid) []func(gp.Population)stats.Plot {
    colors := []string{"#ff8080", "#80ff80", "#8080ff"}
    rows, cols := len(g.colors), len(g.colors[0])
    fn := make([]func(gp.Population)stats.Plot, 3)
    for i := range fn {
        color := i
        fn[color] = func(pop gp.Population) stats.Plot {
            grid := run(g, pop.Best().Code)
            plot := stats.NewPlot("", 0)
            plot.Bubbles.Show = true
            plot.Bubbles.Type = "box"
            plot.Color = colors[color]
            for y := 0; y < cols; y++ {
                for x := 0;  x < rows; x++ {
                    if grid.color(y,x) == color {
                        plot.Data = append(plot.Data, [3]float64{ float64(x), float64(rows-y), 1 })
                    }
                }
            }
            plot.Data = append(plot.Data, [3]float64{-0.5, 0.5, 0.01})
            plot.Data = append(plot.Data, [3]float64{float64(rows)-0.5, float64(cols)+0.5, 0.01})
            return plot
        }
    }
    return fn
}

// function to plot ants and colors they are carrying
func plotAnts(g *Grid) []func(gp.Population)stats.Plot {
    rows := len(g.colors)
    colors := []string{"#000000", "#ff0000", "#00ff00", "#0000ff"}
    fn := make([]func(gp.Population)stats.Plot, 4)
    for i := range fn {
        color := i
        fn[color] = func(pop gp.Population) stats.Plot {
            grid := run(g, pop.Best().Code)
            plot := stats.NewPlot("", 0)
            plot.Bubbles.Show = true
            plot.Bubbles.Fill = color > 0
            plot.Color = colors[color]
            for _, ant := range grid.ants {
                if color == 0 || ant.carrying == color-1 {
                    plot.Data = append(plot.Data, [3]float64{ float64(ant.col), float64(rows-ant.row), 0.8 })                    
                }
            }
            return plot
        }
    }
    return fn
}

type genProxy struct { expr gp.Expr }

func (g genProxy) String() string { return "genProxy" }

func (g genProxy) Generate() *gp.Individual {
    return &gp.Individual{Code: g.expr}
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
    pset.Add(X, Y, CARRYING, COLOR, GO_N, GO_E, GO_S, GO_W, GO_RAND, 
             PICKUP, IFLTE, IFLTZ, IFDROP)

    // setup model
    // gen := gp.GenFull(pset, 1, 2)
    gen := genProxy { gp.Expr{ IFLTE, GO_W, IFDROP, GO_N, IFLTE, GO_RAND, COLOR, Y, COLOR,
                               IFLTE, X, COLOR, COLOR, IFDROP, CARRYING, PICKUP, GO_RAND } }

    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gen,
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

    logger := &stats.Logger{ MaxGen: opts.MaxGen, TargetFitness: opts.TargetFitness }
    if opts.Verbose {
        logger.OnStep = func(best *gp.Individual) {
            g := run(grid, best.Code)
            fmt.Println(g)     
        }
    }

    // run
    if opts.Plot {
        logger.RegisterPlot(plotGrid(grid)...)
        logger.RegisterPlot(plotAnts(grid)...) 
        go stats.MainLoop(problem, logger)
        stats.StartBrowser("http://localhost:8080")
        logger.ListenAndServe(":8080", "../web")
    } else {
        fmt.Println()
        logger.PrintStats = true
        logger.PrintBest = opts.Verbose
        problem.Run(logger)
    }
}
