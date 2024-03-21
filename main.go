package main

import (
	"os"

	lilith "github.com/pre-eth/lilith/src"
)

func main() {
	lilith.ArgParse(len(os.Args), os.Args)
}
