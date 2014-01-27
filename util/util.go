// Package util provides utility functions for building gogp models.
package util
import (
    "fmt"
    "os"
    "bufio"
    "flag"
    "runtime"
    "bytes"
    "encoding/json"
    "github.com/ajstarks/svgo"
    "github.com/jnb666/gogp/gp"
)

// Options struct holds global configuration options
type Options struct {
    TournSize, MaxGen int
    PopSize, Threads int
    TargetFitness, CrossoverProb, MutateProb float64
    Plot, Verbose bool
    Seed int64
}

var DefaultOptions = Options{
    MaxGen: 40,
    PopSize: 500,
    TournSize: 7,
    TargetFitness: 0.9999,
    Threads: runtime.NumCPU(),
    CrossoverProb: 0.5,
    MutateProb: 0.2,
}

// ParseFlags reads command flags and sets no. of threads and random seed.
func ParseFlags(opts *Options) {
    flag.IntVar(&opts.MaxGen, "gens", opts.MaxGen, "maximum no. of generations")
    flag.IntVar(&opts.TournSize, "tournsize", opts.TournSize, "tournament size")
    flag.IntVar(&opts.PopSize, "popsize", opts.PopSize, "population size")
    flag.IntVar(&opts.Threads, "threads", opts.Threads, "number of parallel threads")
    flag.Int64Var(&opts.Seed, "seed", opts.Seed, "random seed - set randomly if <= 0")
    flag.Float64Var(&opts.TargetFitness, "target", opts.TargetFitness, "target fitness")
    flag.Float64Var(&opts.CrossoverProb, "cxprob", opts.CrossoverProb, "crossover probability")
    flag.Float64Var(&opts.MutateProb, "mutprob", opts.MutateProb, "mutation probability")
    flag.BoolVar(&opts.Plot, "plot", opts.Plot, "serve plot data via http")
    flag.BoolVar(&opts.Verbose, "v", opts.Verbose, "print out best individual so far")
    flag.Parse()
    gp.SetSeed(opts.Seed)
    runtime.GOMAXPROCS(opts.Threads)
}

// Open opens a file for reading and returns line scanner
func Open(path string) *bufio.Scanner {
    file, err := os.Open(path)
    CheckErr(err)
    return bufio.NewScanner(file)    
}

// Read reads a line from the scanner, returns true if line was read
func Read(s *bufio.Scanner, args ...interface{}) bool {
    if s.Scan() {
        _, err := fmt.Sscan(s.Text(), args...)
        CheckErr(err)
        return true
    }
    return false
}

// CheckErr exits with an error if err is not nil
func CheckErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

// positive modulus
func Mod(a, b int) int {
    return (a % b + b) % b
}

// Plot structure represets a svgo plot
type Plot struct {
    *svg.SVG
    buf bytes.Buffer
    cell int
}

// SVGPlot function creates a new plot with given size
func SVGPlot(width, height, cellSize int) *Plot {
    p := new(Plot)
    p.SVG = svg.New(&p.buf)
    p.cell = cellSize
    p.Start(width, height)
    return p
}

// Data method returns the SVG data as a byte array
func (p *Plot) Data() []byte {
    p.End()
    return p.buf.Bytes()
}

// Grid method draws a grid with given size, callback function returns style for each cell
func (p *Plot) AddGrid(cols, rows, delay int, style func(x, y int) string) {
    script := fmt.Sprintf("var cols = %d;\nvar cell = %d;\nvar delay = %d;\n",
                cols, p.cell, delay)
    script += `
function setPos(cells, pt, style, multi) {
    var el = document.getElementById(pt.Id);
    var circle = el.getElementsByTagName("circle")[0];
    circle.setAttribute("cx", cell/2+cell*pt.X);
    circle.setAttribute("cy", cell/2+cell*pt.Y);
    var pos = pt.X + cols*pt.Y;
    if (multi) {
        if (pt.C > 0) {         // pick up
            circle.setAttribute("style", style[pt.C] + ";stroke:white")
            cells[pos].setAttribute("style", style[0])
        } else if (pt.C < 0) {  //drop
            circle.setAttribute("style", "fill:none;stroke:white")
            cells[pos].setAttribute("style", style[-pt.C])         
        }
    } else {
        var s = cells[pos].getAttribute("style");
        if (style[s]) {
            cells[pos].setAttribute("style", style[s]);
        }
    }
}
function animate(cells, path, i, style, multi) {
    setPos(cells, path[i], style, multi);
    if (i+1 < path.length) {
        setTimeout(function (){ animate(cells, path, i+1, style, multi) }, delay);
    }
}
function draw(path, style, multi) {
    var grid = document.getElementById("grid");
    var cells = grid.getElementsByTagName("rect");
    if (typeof running != 'undefined' && running) {
        for (var i=0; i<path.length; i++) {
            setPos(cells, path[i], style, multi)
        }
    } else {
        animate(cells, path, 0, style, multi);
    }
}`
    p.Script("application/javascript", script)
    p.Gid("grid")
    for y := 0; y < rows; y++ {
        for x := 0;  x < cols; x++ {
            p.Square(x*p.cell, y*p.cell, p.cell-1, style(x,y))
        }
    }
    p.Gend()
}

// Circle method draws circle at given location
func (p *Plot) AddCircle(id string, x int, y int, style string) {
    p.Gid(id)
    p.Circle(x*p.cell+p.cell/2, y*p.cell+p.cell/2, int(0.4*float64(p.cell)), style)
    p.Gend()
}

type pt struct {
    Id string
    X, Y int
}

// Animate method moves the given object leaving a trail
func (p *Plot) Animate(id string, path [][2]int, style map[string]string) {
    points := make([]pt, len(path))
    for i, p := range path {
        points[i] = pt{ Id:id, X:p[0], Y:p[1] }
    }
    data, _  := json.Marshal(points)
    sdata, _ := json.Marshal(style)
    script := fmt.Sprintf("var path = %s;\ndraw(path, %s);\n", data, sdata)
    p.Script("application/javascript", script)
}

type pt2 struct {
    Id string
    X, Y, C int
}

// AnimateMulti method moves a set of objects
func (p *Plot) AnimateMulti(idBase string, path [][4]int, styles []string) {
    points := make([]pt2, len(path))
    for i, p := range path {
        id := fmt.Sprintf("%s%d", idBase, p[0])
        points[i] = pt2{ Id:id, X:p[1], Y:p[2], C:p[3] }
    }
    data, _  := json.Marshal(points)
    sdata, _ := json.Marshal(styles)
    script := fmt.Sprintf("var path = %s;\ndraw(path, %s, true);\n", data, sdata)
    p.Script("application/javascript", script)
}

