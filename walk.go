package main

import (
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"strings"
)

func stateMatch(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil {
		return false
	}
	formOk := fd.Name.Name != "" && fd.Recv == nil && len(fd.Type.Params.List) == 1 && len(fd.Type.Results.List) == 1
	if !formOk {
		return false
	}
	res := helperMatch(fd)
	if res && *machine != "" {
		switch t := fd.Type.Params.List[0].Type.(type) {
		case *ast.Ident:
			return strings.Contains(t.Name, *machine)
		case *ast.StarExpr:
			return strings.Contains(t.X.(*ast.Ident).Name, *machine)
		default:
			return false
		}
	}
	return res
}

func helperMatch(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
		return false
	}
	switch rt := fd.Type.Results.List[0].Type.(type) {
	case *ast.Ident: // Simple type.
		return rt.Name == *stateFn
	case *ast.IndexExpr: // Parameterized type.
		return rt.X.(*ast.Ident).Name == *stateFn
	default:
		return false
	}
}

type builder struct {
	fStates []*stateInfo
	helpers map[string]*helper
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
	inputs      []string

	lastRecv string
}

func (si *stateInfo) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Receive statements - inputs.
	if ue, ok := node.(*ast.UnaryExpr); ok && ue.Op == token.ARROW {
		handled := false
		switch expr := ue.X.(type) {
		case *ast.SelectorExpr:
			si.lastRecv = expr.Sel.Name
			handled = true
		case *ast.CallExpr:
			if fs, ok := expr.Fun.(*ast.SelectorExpr); ok {
				si.lastRecv = "func " + fs.Sel.Name
				handled = true
			}
		case *ast.Ident:
			si.lastRecv = expr.Name
			handled = true
		}
		if !handled {
			log.Println("don't know how to deal with an input type", reflect.TypeOf(ue.X))
		}
	}

	// Return statements - transitions.
	if rs, ok := node.(*ast.ReturnStmt); ok {
		si.transitions = append(si.transitions, rs)
		si.inputs = append(si.inputs, si.lastRecv)
		si.lastRecv = ""
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
