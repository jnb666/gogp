package main
// artifical ant example

import (
    "os"
    "bufio"
    "fmt"
    "runtime"
    "flag"
    "github.com/jnb666/gogp/gp"
    "github.com/jnb666/gogp/stats"
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
}

// positive modulus
func mod(a, b int) int {
    return (a % b + b) % b
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
            ant.dir = mod(ant.dir + dir, 4)
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
    nRow = mod(row + []int{1, 0, -1, 0}[dir], rows)
    nCol = mod(col + []int{0, 1, 0, -1}[dir], cols)
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

func checkErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

// read the trail file to setup the grid
func readTrail(file string) *Config {
    f, err := os.Open(file)
    checkErr(err)
    conf := Config{ grid: Grid{} }
    scanner := bufio.NewScanner(f)
    // first line should have max no. of moves
    scanner.Scan()
    _, err = fmt.Sscan(scanner.Text(), &conf.maxMoves, &conf.plotRows, &conf.plotCols)
    fmt.Println("max moves =", conf.maxMoves, "plot size =", conf.plotRows, conf.plotCols)
    checkErr(err)
    // read the grid
    row := 0
    for scanner.Scan() {
        line := scanner.Bytes()
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
    f.Close()
    return &conf
}

// create a new ant - make deep copy of grid
func newAnt(conf *Config) *Ant {
    grid := conf.grid.Clone()
    return &Ant{
        maxMoves: conf.maxMoves,
        row: conf.startRow,
        col: conf.startCol,
        dir: conf.startDir,
        grid: grid,
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

// new bubble plot
func createPlot(label string, grid Grid, rows, cols int, cellType byte, size float64) stats.Plot { 
    plot := stats.NewPlot(label, rows*cols)
    plot.Bubbles.Show = true
    i := 0
    for y := 0; y < cols; y++ {
        for x := 0;  x < rows; x++ {
            if grid[y][x] == cellType {
                plot.Data[i] = [3]float64{ float64(x), float64(rows-y), size }
                i++
            }
        }
    }
    plot.Data = plot.Data[:i]
    return plot
}

// function to plot grid
func plotGrid(c *Config) func(gp.Population) stats.Plot {
    return func(pop gp.Population) stats.Plot {
        plot := createPlot("food", c.grid, c.plotRows, c.plotCols, FOOD, 1.1)
        plot.Color = "#00ff00"
        plot.Bubbles.Type = "box"
        plot.Data = append(plot.Data, [3]float64{-0.5, 0.5, 0.01})
        plot.Data = append(plot.Data, [3]float64{float64(c.plotRows)-0.5, float64(c.plotCols)+0.5, 0.01})
        return plot
    }
}

// function to plot path of best individual
func plotBest(c *Config) func(gp.Population) stats.Plot {
    return func(pop gp.Population) stats.Plot {
        ant := run(c, pop.Best().Code)
        plot := createPlot("best", ant.grid, c.plotRows, c.plotCols, TRAIL, 0.8)
        plot.Color = "#ff0000"
        return plot
    }
}

// build and run model
func main() {
    var trailFile string
    var threads, generations, popsize, depth, size int
    var seed int64
    var plot, verbose bool
    flag.StringVar(&trailFile, "trail", "santafe_trail.txt", "trail definition file")
    flag.IntVar(&threads, "threads", runtime.NumCPU(), "number of parallel threads")
    flag.Int64Var(&seed, "seed", 0, "random seed - set randomly if <= 0")
    flag.IntVar(&generations, "gens", 40, "maximum no. of generations")
    flag.IntVar(&popsize, "popsize", 4000, "population size")
    flag.IntVar(&depth, "depth", 0, "depth limit - or zero for none")
    flag.IntVar(&size, "size", 0, "size limit - or zero for none")
    flag.BoolVar(&plot, "plot", false, "connect to gogpweb to plot statistics")
    flag.BoolVar(&verbose, "v", false, "print out best individual so far")
    flag.Parse()

    gp.SetSeed(seed)
    runtime.GOMAXPROCS(threads)
    config := readTrail(trailFile)

    pset := gp.CreatePrimSet(0)
    pset.Add(progN{ &gp.BaseFunc{"prog2", 2} })
    pset.Add(progN{ &gp.BaseFunc{"prog3", 3} })
    pset.Add(ifFood{ &gp.BaseFunc{"if_food", 2} })
    pset.Add(Terminal("left", turn(-1)))
    pset.Add(Terminal("right", turn(1)))
    pset.Add(Terminal("step", step))

    problem := &gp.Model{
        PrimitiveSet: pset,
        Generator: gp.GenFull(pset, 1, 2),
        PopSize: popsize,
        Fitness: fitnessFunc(config),
        Offspring: gp.Tournament(7),
        Mutate: gp.MutUniform(gp.GenFull(pset, 0, 2)),
        MutateProb: 0.2,
        Crossover: gp.CxOnePoint(),
        CrossoverProb: 0.5,
        Threads: threads,
    }
    if depth > 0 {
        problem.AddDecorator(gp.DepthLimit(depth))
    }
    if size > 0 {
        problem.AddDecorator(gp.SizeLimit(size))
    }

    problem.PrintParams("== Artificial ant ==")
    fmt.Println()

    logger := &stats.Logger{ MaxGen: generations, TargetFitness: 0.999 }
    if verbose {
        logger.OnDone = func(best *gp.Individual) {
            ant := run(config, best.Code)
            fmt.Println(ant.grid)            
        }
    }

    if plot {
        logger.RegisterPlot(plotGrid(config)) 
        logger.RegisterPlot(plotBest(config)) 
        go stats.MainLoop(problem, logger)
        stats.StartBrowser("http://localhost:8080")
        logger.ListenAndServe(":8080", "../web")
    } else {
        logger.PrintStats = true
        logger.PrintBest = verbose
        problem.Run(logger)
    }
}

