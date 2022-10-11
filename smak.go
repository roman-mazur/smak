// Command smak can be used to visualize a state machine implemented using the state function approach.
// See https://www.youtube.com/watch?v=HxaD_trXwRE
package main

import (
	"flag"
	"go/ast"
	"log"
	"os"
	"reflect"
)

var (
	stateFn = flag.String("state-fn", "stateFn", "State function type name")
	out     = flag.String("out", "edges", "Type of the output")
)

func main() {
	flag.Parse()

	fileName := flag.Arg(0)
	if fileName == "" {
		log.Fatal("Input path is required")
	}

	fnode, err := parse(fileName)
	if err != nil {
		log.Fatal("Failed to parse the input file: ", err)
	}

	var b builder
	ast.Walk(&b, fnode)

	switch *out {
	case "matrix":
		_, err = fprintMatrixDot(os.Stdout, *stateFn, states(&b), matrix(&b))
	case "edges":
		_, err = fprintEdgesDot(os.Stdout, *stateFn, edges(&b))
	}

	if err != nil {
		panic(err)
	}
}

func states(b *builder) []string {
	res := make([]string, len(b.fStates)+1)
	for i := range b.fStates {
		res[i] = b.fStates[i].state.Name.Name
	}
	res[len(res)-1] = "_terminated_"
	return res
}

func matrix(b *builder) [][]bool {
	n := len(b.fStates) + 1
	idx := make(map[string]int, n)
	for i := range b.fStates {
		idx[b.fStates[i].state.Name.Name] = i
	}
	idx["nil"] = len(idx)

	res := make([][]bool, n)
	for i := range b.fStates {
		res[i] = make([]bool, n)
		for _, node := range b.fStates[i].transitions {
			for _, d := range transitionNames(b, node) {
				if j, exists := idx[d]; exists {
					res[i][j] = true
				} else {
					panic("bad destination " + d + " from " + reflect.TypeOf(node.Results[0]).String())
				}
			}
		}
	}
	return res
}

type edge struct {
	from, to, label string
}

func edges(b *builder) []edge {
	var res []edge
	for _, s := range b.fStates {
		for i, node := range s.transitions {
			for _, d := range transitionNames(b, node) {
				res = append(res, edge{
					from:  s.state.Name.Name,
					to:    d,
					label: s.inputs[i],
				})
			}
		}
	}
	return res
}

func transitionNames(b *builder, node *ast.ReturnStmt) []string {
	var dst []string
	switch tr := node.Results[0].(type) {
	case *ast.Ident:
		dst = []string{tr.Name}
	case *ast.CallExpr:
		if sel, ok := tr.Fun.(*ast.SelectorExpr); ok {
			if h, found := b.helpers[sel.Sel.Name]; found {
				dst = h.targets
			} else {
				panic("call to not a helper " + sel.Sel.Name)
			}
		} else {
			panic("unknown call Fun type " + reflect.TypeOf(tr.Fun).String())
		}
	default:
		panic("don't know what to do with " + reflect.TypeOf(node.Results[0]).String())
	}
	return dst
}
