package main

// painted desert example
// from Evaluation of Emergent Cooperative Behaviour using Genetic Progamming [Koza 1993]

import (
	"flag"
	"fmt"
	"github.com/jnb666/gogp/gp"
	"github.com/jnb666/gogp/stats"
	"github.com/jnb666/gogp/util"
	"math/rand"
	"strings"
)

// colour mappings
var (
	Colors   = []byte{'.', 'R', 'G', 'B'}
	ColorsLC = []byte{'o', 'r', 'g', 'b'}
)

type Cell struct {
	color int
	ant   bool
}

// grid of cells and other global vars
type Grid struct {
	cells           [][]Cell
	rng             *rand.Rand
	rows, cols      int
	steps, maxMoves int
	path            [][4]int
}

// per ant data
type Ant struct {
	id, row, col, carrying, moves int
	pickupRow, pickupCol          int
	pickupFlag                    bool
	grid                          *Grid
}

// print out the grid
func (g *Grid) String() string {
	lines := make([]string, g.rows)
	for y, row := range g.cells {
		for _, cell := range row {
			if cell.ant {
				lines[y] += string(ColorsLC[cell.color+1])
			} else {
				lines[y] += string(Colors[cell.color+1])
			}
		}
	}
	return strings.Join(lines, "\n")
}

// deep copy of grid data
func (g *Grid) clone() *Grid {
	grid := *g
	grid.cells = make([][]Cell, g.rows)
	for row, line := range g.cells {
		grid.cells[row] = append([]Cell{}, line...)
	}
	grid.rng = rand.New(rand.NewSource(1))
	return &grid
}

// get next position
func (g *Grid) next(row, col, dir int) (nRow, nCol int) {
	nRow = util.Mod(row+[]int{1, 0, -1, 0}[dir], g.rows)
	nCol = util.Mod(col+[]int{0, 1, 0, -1}[dir], g.cols)
	return
}

// get initial ants
func (g *Grid) ants() []*Ant {
	ants := []*Ant{}
	id := 0
	for row := 0; row < g.rows; row++ {
		for col := 0; col < g.cols; col++ {
			if g.cells[row][col].ant {
				ants = append(ants, &Ant{id: id, row: row, col: col, carrying: -1, grid: g})
				id++
			}
		}
	}
	return ants
}

// read the trail file to setup the grid
func readGrid(file string) *Grid {
	s := util.Open(file)
	grid := Grid{}
	// first line has config params
	util.Read(s, &grid.steps, &grid.maxMoves)
	fmt.Printf("steps=%d maxMoves=%d\n", grid.steps, grid.maxMoves)
	// read the initial grid
	grid.cells = [][]Cell{}
	row := 0
	for s.Scan() {
		line := s.Bytes()
		grid.cols = len(line)
		grid.cells = append(grid.cells, make([]Cell, grid.cols))
		for col, cell := range line {
			color := -1
			switch cell {
			case 'R', 'r':
				color = 0
			case 'G', 'g':
				color = 1
			case 'B', 'b':
				color = 2
			}
			grid.cells[row][col].color = color
			if cell >= 'a' && cell <= 'z' {
				grid.cells[row][col].ant = true
			}
		}
		row++
	}
	grid.rows = row
	return &grid
}

// terminal set
type terminal struct {
	*gp.BaseFunc
	fn func(*Ant) int
}

func Terminal(name string, fn func(*Ant) int) gp.Opcode {
	return terminal{&gp.BaseFunc{name, 0}, fn}
}

func (o terminal) Eval(args ...gp.Value) gp.Value {
	return o.fn
}

// move in given direction
func move(dir int) func(*Ant) int {
	return func(ant *Ant) int {
		if ant.moves < ant.grid.maxMoves {
			ant.moves++
			row, col := ant.grid.next(ant.row, ant.col, dir)
			if !ant.grid.cells[row][col].ant {
				ant.grid.cells[ant.row][ant.col].ant = false
				ant.grid.cells[row][col].ant = true
				ant.row, ant.col = row, col
				// update board if we have just picked up a grain
				if ant.grid.path != nil && ant.pickupFlag {
					ant.grid.path = append(ant.grid.path,
						[4]int{ant.id, ant.pickupCol, ant.pickupRow, ant.carrying + 1})
					ant.pickupFlag = false
				}
			}
		}
		return ant.grid.cells[ant.row][ant.col].color
	}
}

// pick up grain at current pos if not currently carrying anything
func pickUp(ant *Ant) int {
	color := ant.grid.cells[ant.row][ant.col].color
	if ant.moves < ant.grid.maxMoves && ant.carrying < 0 && color >= 0 {
		ant.moves++
		ant.carrying = color
		ant.pickupRow, ant.pickupCol = ant.row, ant.col
		ant.grid.cells[ant.row][ant.col].color = -1
		ant.pickupFlag = true
		return -1
	}
	return color
}

// function set
type ifelse struct {
	*gp.BaseFunc
	cond func(*Ant, ...gp.Value) bool
}

func IfElse(name string, arity int, cond func(*Ant, ...gp.Value) bool) gp.Opcode {
	return ifelse{&gp.BaseFunc{name, arity}, cond}
}

func (o ifelse) Eval(arg ...gp.Value) gp.Value {
	arity := o.Arity()
	return func(ant *Ant) int {
		if o.cond(ant, arg...) {
			return arg[arity-2].(func(*Ant) int)(ant)
		} else {
			return arg[arity-1].(func(*Ant) int)(ant)
		}
	}
}

// if arg0 <= arg1 then arg2 else arg3
func ifLessThanOrEqual(ant *Ant, arg ...gp.Value) bool {
	return arg[0].(func(*Ant) int)(ant) <= arg[1].(func(*Ant) int)(ant)
}

// if arg0 < 0 then arg1 else arg2
func ifLessThanZero(ant *Ant, arg ...gp.Value) bool {
	return arg[0].(func(*Ant) int)(ant) < 0
}

// if carrying a grain and current position is empty drop and call arg0, else call arg1
func ifDrop(ant *Ant, arg ...gp.Value) bool {
	if ant.carrying >= 0 && ant.grid.cells[ant.row][ant.col].color < 0 {
		if ant.moves < ant.grid.maxMoves {
			ant.grid.cells[ant.row][ant.col].color = ant.carrying
			// update board if dropped in new location
			if ant.grid.path != nil && !ant.pickupFlag {
				ant.grid.path = append(ant.grid.path,
					[4]int{ant.id, ant.col, ant.row, -ant.carrying - 1})
			}
			ant.moves++
			ant.carrying = -1
		}
		return true
	}
	return false
}

// run the code - step each ant in turn
func run(g *Grid, code gp.Expr) (*Grid, [][][4]int) {
	runFunc := code.Eval().(func(*Ant) int)
	grid := g.clone()
	ants := grid.ants()
	pathList := make([][][4]int, 0, grid.steps)
	for i := 0; i < grid.steps; i++ {
		grid.path = make([][4]int, 0, len(ants))
		for _, ant := range ants {
			if ant.moves < grid.maxMoves {
				row, col := ant.row, ant.col
				runFunc(ant)
				if ant.col != col || ant.row != row {
					grid.path = append(grid.path, [4]int{ant.id, ant.col, ant.row, 0})
				}
			}
		}
		if len(grid.path) > 0 {
			pathList = append(pathList, grid.path)
		}
	}
	grid.path = make([][4]int, 0, len(ants))
	finalise(ants)
	if len(grid.path) > 0 {
		pathList = append(pathList, grid.path)
	}
	return grid, pathList
}

// move to given location and count grain there if it is empty
func drop(ant *Ant, row, col int) bool {
	if ant.carrying >= 0 && ant.grid.cells[row][col].color < 0 {
		ant.grid.cells[row][col].color = ant.carrying
		if ant.grid.path != nil && (row != ant.row || col != ant.col) {
			ant.grid.path = append(ant.grid.path, [4]int{ant.id, col, row, 0})
		}
		return true
	}
	return false
}

// if ant is holding a grain it must drop it so that it can be counted
func finalise(ants []*Ant) {
Loop:
	for _, ant := range ants {
		if ant.carrying < 0 {
			continue
		}
		// if current location is empty, count it here
		if drop(ant, ant.row, ant.col) {
			continue
		}
		// if cell we picked it up from is empty, count it there
		if drop(ant, ant.pickupRow, ant.pickupCol) {
			continue
		}
		// find next free space
		for xoff := 1; xoff < ant.grid.cols; xoff++ {
			x := util.Mod(ant.col+xoff, ant.grid.cols)
			for yoff := 0; yoff <= xoff; yoff++ {
				y := util.Mod(ant.row+yoff, ant.grid.rows)
				if drop(ant, y, x) {
					continue Loop
				}
				if yoff > 0 {
					y := util.Mod(ant.row-yoff, ant.grid.rows)
					if drop(ant, y, x) {
						continue Loop
					}
				}
			}
		}
		panic("nowhere to drop the sand!")
	}
}

// Raw fitness is the product of the color of the grain of sand (1,2,3) and the distance
// between the grain and the Y-axis when execution of the particular program ceases.
func fitnessFunc(g *Grid) func(gp.Expr) (float64, bool) {
	return func(code gp.Expr) (float64, bool) {
		runFunc := code.Eval().(func(*Ant) int)
		grid := g.clone()
		ants := grid.ants()
		for i := 0; i < grid.steps; i++ {
			for _, ant := range ants {
				if ant.moves < grid.maxMoves {
					runFunc(ant)
				}
			}
		}
		finalise(ants)
		fit := 0
		// check every square for sand on the ground
		for row, line := range grid.cells {
			for col := range line {
				switch grid.cells[row][col].color {
				case 0:
					fit += 3 * (col + 1)
				case 1:
					fit += 2 * (col + 1)
				case 2:
					fit += col + 1
				}
			}
		}
		// scale fitness to 0 to 1 range
		return 100 / float64(fit), true
	}
}

// returns function to plot path of best individual
func createPlot(g *Grid, size, delay int) func(gp.Population) []byte {
	styles := []string{"fill:grey", "fill:red", "fill:green", "fill:blue"}
	sz := size / g.cols
	return func(pop gp.Population) []byte {
		ch := make(chan [][][4]int)
		go func() {
			_, path := run(g, pop.Best().Code)
			ch <- path
		}()
		// draw grid
		plot := util.SVGPlot(size, size, sz)
		plot.AddGrid(g.cols, g.rows, delay, func(x, y int) string {
			return styles[g.cells[y][x].color+1]
		})
		// draw ants
		plot.Gid("ant")
		for _, ant := range g.ants() {
			plot.Circle(ant.col*sz+sz/2, ant.row*sz+sz/2, int(0.4*float64(sz)), "fill:none")
		}
		plot.Gend()
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
	pset.Add(Terminal("x", func(ant *Ant) int { return ant.col }))
	pset.Add(Terminal("y", func(ant *Ant) int { return ant.row }))
	pset.Add(Terminal("carrying", func(ant *Ant) int { return ant.carrying }))
	pset.Add(Terminal("color", func(ant *Ant) int { return ant.grid.cells[ant.row][ant.col].color }))
	pset.Add(Terminal("go-n", move(0)))
	pset.Add(Terminal("go-e", move(1)))
	pset.Add(Terminal("go-s", move(2)))
	pset.Add(Terminal("go-w", move(3)))
	pset.Add(Terminal("go-rand", func(ant *Ant) int { return move(ant.grid.rng.Intn(4))(ant) }))
	pset.Add(Terminal("pickup", pickUp))
	pset.Add(IfElse("iflte", 4, ifLessThanOrEqual))
	pset.Add(IfElse("ifltz", 3, ifLessThanZero))
	pset.Add(IfElse("ifdrop", 2, ifDrop))

	// setup model
	problem := &gp.Model{
		PrimitiveSet:  pset,
		Generator:     gp.GenFull(pset, 1, 2),
		PopSize:       opts.PopSize,
		Fitness:       fitnessFunc(grid),
		Offspring:     gp.Tournament(opts.TournSize),
		Mutate:        gp.MutUniform(gp.GenFull(pset, 0, 2)),
		MutateProb:    opts.MutateProb,
		Crossover:     gp.CxOnePoint(),
		CrossoverProb: opts.CrossoverProb,
		Threads:       opts.Threads,
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
			g, _ := run(grid, best.Code)
			fmt.Println(g)
		}
	}

	// run
	if opts.Plot {
		logger.RegisterSVGPlot("best", createPlot(grid, 500, 40))
		stats.MainLoop(problem, logger, ":8080", "../web")
	} else {
		fmt.Println()
		logger.PrintStats = true
		logger.PrintBest = opts.Verbose
		problem.Run(logger)
	}
}
