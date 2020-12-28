package main

import (
	"os"
	"runtime"

	"github.com/k-kinzal/progressived/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := cmd.Execute(); err != nil {
		os.Exit(255)
	}
}
