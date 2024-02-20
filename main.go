package main

import (
	"time"
	"crypto/rand"
    "bytes"
    "encoding/binary"
	"github.com/pre-eth/lilith/lilith"
)

func main() {
	bytebuf := make([]byte, 32)
	rand.Read(bytebuf[0:31])

	seed := make([]uint16, 8)
	rbuf := bytes.NewReader(bytebuf[0:16])
    binary.Read(rbuf, binary.LittleEndian, &seed)

	nonce, _ := binary.Uvarint(bytebuf[16:24])

	salt, _ := binary.Uvarint(bytebuf[24:32])
	currTime := time.Now().Unix()
	salt ^= uint64(currTime) << (currTime & 31)

	lilith.Lilith(seed, nonce, salt)
}
