package main

import (
	"fmt"
	"io"
)

func fprintDot(out io.Writer, name string, states []string, matrix [][]bool) (n int, err error) {
	printIt := func(format string, args ...any) {
		if err != nil {
			return
		}
		nn, ee := fmt.Fprintf(out, format, args...)
		n += nn
		err = ee
	}

	printIt("digraph %s {\n", name)
	for i, row := range matrix {
		for j := range row {
			if matrix[i][j] {
				printIt("  %s -> %s;\n", states[i], states[j])
			}
		}
	}
	printIt("}\n")
	return
}
