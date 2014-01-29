package main
// artifical ant example

import (
    "fmt"
    "flag"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
    "github.com/jnb666/gogp/util"
)

const (
    FOOD  = '#'
    TRAIL = '*'
    START = 'S'
)

// grid of cells
type Grid [][]byte

// global config data
type Config struct {
    startRow, startCol, startDir int
    maxMoves, totalFood int
    plotRows, plotCols int
    grid Grid
}

// ant data
type Ant struct {
    row, col, dir int
    maxMoves, moves, eaten int
    grid Grid
    path [][2]int
}

// execute each of args in sequence
type progN struct { *gp.BaseFunc }

func (o progN) Eval(args ...gp.Value) gp.Value {
    return func(ant *Ant) {
        for _, arg := range args {
            arg.(func(*Ant))(ant)
        }
    }
}

// terminal node
type terminal struct {
    *gp.BaseFunc
    fn func(*Ant)
}

func Terminal(name string, fn func(*Ant)) gp.Opcode {
    return terminal{&gp.BaseFunc{name,0}, fn}
}

func (o terminal) Eval(args ...gp.Value) gp.Value {
    return o.fn
}

// turn left or left
func turn(dir int) func(*Ant) {
    return func(ant *Ant) {
        if ant.moves < ant.maxMoves {
            ant.moves++
            ant.dir = util.Mod(ant.dir + dir, 4)
            ant.grid[ant.row][ant.col] = TRAIL
        }
    }
}

// step forward and pick up food if in new cell
func step(ant *Ant) {
    if ant.moves < ant.maxMoves {
        ant.moves++
        ant.row, ant.col = ant.grid.Next(ant.row, ant.col, ant.dir)
        if ant.grid[ant.row][ant.col] == FOOD { ant.eaten++ }
        ant.grid[ant.row][ant.col] = TRAIL
        ant.path = append(ant.path, [2]int{ant.col, ant.row})
    }
}

// if next cell contains food execute first arg else execute second
type ifFood struct { *gp.BaseFunc }

func (o ifFood) Eval(args ...gp.Value) gp.Value {
    return func(ant *Ant) {
        row, col := ant.grid.Next(ant.row, ant.col, ant.dir)
        if ant.grid[row][col] == FOOD {
            args[0].(func(*Ant))(ant)
        } else {
            args[1].(func(*Ant))(ant)
        }
    }
}

// grid methods
func (g Grid) Next(row, col, dir int) (nRow, nCol int) {
    rows, cols := len(g), len(g[0])
    nRow = util.Mod(row + []int{1, 0, -1, 0}[dir], rows)
    nCol = util.Mod(col + []int{0, 1, 0, -1}[dir], cols)
    return
}

func (g Grid) String() (text string) {
    for _, line := range g {
        text += string(line) + "\n"
    }
    return
}

func (g Grid) Clone() Grid {
    grid := make(Grid, len(g))
    for row, line := range g {
        grid[row] = append([]byte{}, line...)
    }
    return grid
}

// read the trail file to setup the grid
func readTrail(file string) *Config {
    s := util.Open(file)
    conf := Config{ grid: Grid{} }
    // first line has max no. of moves and plot dimensions
    util.Read(s, &conf.maxMoves, &conf.plotRows, &conf.plotCols)
    fmt.Println("max moves =", conf.maxMoves, "plot size =", conf.plotRows, conf.plotCols)
    // read the grid
    row := 0
    for s.Scan() {
        line := s.Bytes()
        for col, cell := range line {
            switch cell {
            case FOOD:
                conf.totalFood++
            case START:
                conf.startRow, conf.startCol = row, col
                conf.startDir = 1
                line[col] = TRAIL
            }
        }
        copy := append([]byte{}, line...)
        conf.grid = append(conf.grid, copy)
        row++
    }
    return &conf
}

// create a new ant - make deep copy of grid
func newAnt(conf *Config) *Ant {
    grid := conf.grid.Clone()
    path := make([][2]int, 1, conf.maxMoves)
    path[0][0], path[0][1] = conf.startCol, conf.startRow
    return &Ant{
        maxMoves: conf.maxMoves,
        row: conf.startRow,
        col: conf.startCol,
        dir: conf.startDir,
        grid: grid,
        path: path,
    }
}

// run the code
func run(conf *Config, code gp.Expr) *Ant {
    ant := newAnt(conf)
    runFunc := code.Eval().(func(*Ant))
    for ant.moves < ant.maxMoves { runFunc(ant) }
    return ant
}

// run the program to calculate the fitness as no. of food cells eaten / total
func fitnessFunc(conf *Config) func(gp.Expr) (float64, bool) {
    return func(code gp.Expr) (float64, bool) {
        ant := run(conf, code)
        return float64(ant.eaten)/float64(conf.totalFood), true
    }
}

// returns function to plot path of best individual
func createPlot(c *Config, size, delay int) func(gp.Population) []byte {
    sz := size/c.plotCols
    return func(pop gp.Population) []byte {
        ch := make(chan [][2]int)
        go func() {
            ant := run(c, pop.Best().Code)
            ch <- ant.path
        }()
        // draw grid
        plot := util.SVGPlot(size, size, sz)
        plot.AddGrid(c.plotCols, c.plotRows, delay, func(x, y int) string {
            if c.grid[y][x] == FOOD {
                return "fill:green"
            } else {
                return "fill:grey"
            }
        })
        // draw ant
        plot.Gid("ant")
        plot.Circle(c.startCol*sz+sz/2, c.startRow*sz+sz/2, int(0.4*float64(sz)), "fill:black")
        plot.Gend()
        plot.Animate("ant", <-ch, 
            map[string]string{"fill:grey":"fill:brown", "fill:green":"fill:red"})
        return plot.Data()
    }
}

// build and run model
func main() {
    // get options
    var maxSize, maxDepth int
    var trailFile string
    flag.IntVar(&maxSize, "size", 0, "maximum tree size - zero for none")
    flag.IntVar(&maxDepth, "depth", 0, "maximum tree depth - zero for none")
    flag.StringVar(&trailFile, "trail", "santafe_trail.txt", "trail definition file")    
    opts := util.DefaultOptions
    util.ParseFlags(&opts)

    // create primitive set
    config := readTrail(trailFile)
    pset := gp.CreatePrimSet(0)
    pset.Add(progN{ &gp.BaseFunc{"prog2", 2} })
    pset.Add(progN{ &gp.BaseFunc{"prog3", 3} })
    pset.Add(ifFood{ &gp.BaseFunc{"if_food", 2} })
    pset.Add(Terminal("left", turn(-1)))
    pset.Add(Terminal("right", turn(1)))
    pset.Add(Terminal("step", step))

    // setup model
    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 1, 2),
        PopSize: opts.PopSize,
        Fitness: fitnessFunc(config),
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
            ant := run(config, best.Code)
            fmt.Println(ant.grid)
        }
    }

    // run
    if opts.Plot {
        logger.RegisterSVGPlot("best", createPlot(config, 500, 10)) 
        stats.MainLoop(problem, logger, ":8080", "../web")
    } else {
        fmt.Println()
        logger.PrintStats = true
        logger.PrintBest = opts.Verbose
        problem.Run(logger)
    }
}

