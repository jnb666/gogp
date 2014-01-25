package gp
import (
    "fmt"
    "strconv"
    "errors"
    "bufio"
    "bytes"
    "os"
    "os/exec"
    "io/ioutil"
    gv "code.google.com/p/gographviz"
)

// Default attributes for graph plotting
var (
    NodeAttrs = gv.Attrs{
        "fontname": `"Helvetica"`,
        "fontsize": "10",
        "style":    "filled",
        "color":    "lightgrey",
    }
    GraphDPI = "72"
)

// Graph returns a graphiz graph for the expression.
// This depends on the "code.google.com/p/gographviz" module.
func (e Expr) Graph(name string) *gv.Graph {
    g := gv.NewGraph()
    g.SetName(name)
    g.Attrs.Add("dpi", GraphDPI)
    // add nodes
    for i, op := range e {
        attrs := NodeAttrs.Copy()
        attrs.Add("label", `"`+op.String()+`"`)
        g.AddNode(name, strconv.Itoa(i), attrs)
    }
    // recursively find edges starting from top
    var getChild func() string
    pos := 0
    getChild = func() string {
        op := e[pos]
        arity := op.Arity()
        if arity == 0 {
            return strconv.Itoa(pos)
        }
        parent := strconv.Itoa(pos)
        for i:=0; i<arity; i++ {
            pos++
            g.AddEdge(parent, "", getChild(), "", false, nil)
        }
        return parent
    }
    getChild()
    return g
}

// Layout runs the dot program to layout the graph.
// If format is "" uses dot format, else specify an output renderer, e.g. "svg".
// The Graphviz dot program must be installed in your path.
func Layout(graph *gv.Graph, format string) ([]byte, error) {
    f, err := ioutil.TempFile("", "gogp")
    if err != nil {
        return nil, err
    }
    // write to tmpfile
    w := bufio.NewWriter(f)
    fmt.Fprint(w, graph.String())
    w.Flush()
    f.Close()
    file := f.Name()
    defer os.Remove(file)
    // process with dot
    args := []string{ file }
    if format != "" {
        args = append(args, "-T" + format)
    }
    var outData, errData bytes.Buffer
    cmd := exec.Command("dot", args...)
    cmd.Stdout = &outData
    cmd.Stderr = &errData
    if err := cmd.Run(); err != nil {
        if errData.Len() > 0 {
            err = errors.New(errData.String())
        }
        return nil, err
    }
    return outData.Bytes(), nil
}


