package main

import (
	"os"

	"github.com/k-kinzal/progressived/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(255)
	}
}
