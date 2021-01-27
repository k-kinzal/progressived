package main

import (
	"fmt"
	"os"

	"github.com/k-kinzal/progressived/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("progressived: %v", err))
		os.Exit(255)
	}
}
