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
	_, err = fprintDot(os.Stdout, *stateFn, b.States(), b.Matrix())
	if err != nil {
		panic(err)
	}
}

func stateMatch(fd *ast.FuncDecl) bool {
	formOk := fd.Name.Name != "" && fd.Recv == nil && len(fd.Type.Params.List) == 1 && len(fd.Type.Results.List) == 1
	if !formOk {
		return false
	}
	return helperMatch(fd)
}

func helperMatch(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
		return false
	}
	if rt, ok := fd.Type.Results.List[0].Type.(*ast.Ident); ok {
		return rt.Name == *stateFn
	}
	return false
}

type builder struct {
	fStates []*stateInfo
	helpers map[string]*helper
}

func (b *builder) States() []string {
	res := make([]string, len(b.fStates)+1)
	for i := range b.fStates {
		res[i] = b.fStates[i].state.Name.Name
	}
	res[len(res)-1] = "_terminated_"
	return res
}

func (b *builder) Matrix() [][]bool {
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

			for _, d := range dst {
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

func (b *builder) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if fd, ok := node.(*ast.FuncDecl); ok && stateMatch(fd) {
		fi := &stateInfo{state: fd}
		b.fStates = append(b.fStates, fi)
		return fi
	}
	if fd, ok := node.(*ast.FuncDecl); ok && helperMatch(fd) {
		if b.helpers == nil {
			b.helpers = make(map[string]*helper)
		}
		h := &helper{name: fd.Name.Name}
		b.helpers[h.name] = h
		return h
	}
	return b
}

type stateInfo struct {
	state       *ast.FuncDecl
	transitions []*ast.ReturnStmt
}

func (si *stateInfo) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if rs, ok := node.(*ast.ReturnStmt); ok {
		si.transitions = append(si.transitions, rs)
	}
	return si
}

type helper struct {
	name    string
	targets []string
}

func (h *helper) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if rs, ok := node.(*ast.ReturnStmt); ok {
		if id, ok := rs.Results[0].(*ast.Ident); ok {
			h.targets = append(h.targets, id.Name)
		} else {
			panic("it's too complex: helper " + h.name + " has return " + reflect.TypeOf(rs.Results[0]).String())
		}
	}
	return h
}
