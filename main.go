package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"

	"github.com/pre-eth/lilith/lilith"
)

func main() {
	bytebuf := make([]byte, 28)
	rand.Read(bytebuf[0:16])

	seed := make([]uint16, 8)
	rbuf := bytes.NewReader(bytebuf[0:16])
	binary.Read(rbuf, binary.LittleEndian, &seed)

	nonce := make([]uint32, 3)
	rbuf = bytes.NewReader(bytebuf[16:28])
	binary.Read(rbuf, binary.LittleEndian, &nonce)

	ciphertext := make([]byte, 0)

	lilith.Lilith(ciphertext, seed, nonce)
}
