package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/pre-eth/lilith/lilith"
)

const (
	ERR_COLOR = "\033[1;38;2;255;17;0m"
)

func delayedPrint(text string, style string, delay_time float32) {
	delay := 0
	for delay < 0 {
		delay = int(delay_time)
		delay_time *= 10.0
	}

	text_length := len(text)
	i := 0
	char := ""
	fmt.Print(style)
	for i < text_length {
		time.Sleep(time.Duration(delay_time) * time.Millisecond)
		runeValue, w := utf8.DecodeRuneInString(text[i:])
		fmt.Printf("%c", runeValue)
		if char == "." && i != text_length-1 {
			j := 0
			for j < 2 {
				fmt.Print(" \033[1D.")
				time.Sleep(time.Duration(50) * time.Millisecond)
				j += 1
			}
		}
		i += w
	}
	fmt.Print("\033[m\n")
}

func main() {
	argc := len(os.Args)
	if argc == 1 {
		delayedPrint("❌ No file provided. Exiting.", ERR_COLOR, 27)
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
