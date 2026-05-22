package main

import (
	"fmt"
	"os"

	"github.com/yacchin1205/lc4ri/internal/lc4ri"
)

func main() {
	if err := lc4ri.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "lc4ri:", err)
		os.Exit(1)
	}
}
