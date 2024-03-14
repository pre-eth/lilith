package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/pre-eth/lilith/lilith"
)

func main() {
	argc := len(os.Args)
	if argc == 0 {
		fmt.Println("No file provided. Exiting.")
		os.Exit(0)
	}

	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	bytebuf := make([]byte, 28)
	rand.Read(bytebuf[0:16])

	seed := make([]uint16, 8)
	rbuf := bytes.NewReader(bytebuf[0:16])
	binary.Read(rbuf, binary.LittleEndian, &seed)

	nonce := make([]uint32, 3)
	rbuf = bytes.NewReader(bytebuf[16:28])
	binary.Read(rbuf, binary.LittleEndian, &nonce)

	lilith.Lilith(file, seed, nonce)
}
