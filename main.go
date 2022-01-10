package main

import (
	"log"
	"os"

	"github.com/lusingander/edist/internal/edist"
)

func run(args []string) error {
	return edist.Start()
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
