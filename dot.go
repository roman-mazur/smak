package main

import (
	"fmt"
	"io"
)

func fprintMatrixDot(out io.Writer, name string, states []string, matrix [][]bool) (n int, err error) {
	p := dotPrint{out: out}
	p.begin(name)
	for i, row := range matrix {
		for j := range row {
			if matrix[i][j] {
				p.printEdge(states[i], states[j])
			}
		}
	}
	p.end()
	return p.n, p.err
}

func fprintEdgesDot(out io.Writer, name string, edges []edge) (n int, err error) {
	p := dotPrint{out: out}
	p.begin(name)
	for _, e := range edges {
		p.printNamesEdge(e.from, e.to, e.label)
	}
	p.end()
	return p.n, p.err
}

type dotPrint struct {
	n   int
	err error
	out io.Writer
}

func (dp *dotPrint) begin(name string) {
	dp.print("digraph %s {\n", name)
}

func (dp *dotPrint) printEdge(src, dst string) {
	dp.print("  %s -> %s;\n", src, dst)
}

func (dp *dotPrint) printNamesEdge(src, dst, label string) {
	dp.print("  %s -> %s [label=%q];\n", src, dst, label)
}

func (dp *dotPrint) end() {
	dp.print("}\n")
}

func (dp *dotPrint) print(format string, args ...any) {
	if dp.err != nil {
		return
	}
	n, err := fmt.Fprintf(dp.out, format, args...)
	dp.n += n
	dp.err = err
}
