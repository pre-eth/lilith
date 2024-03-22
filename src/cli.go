package lilith

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MAJOR       = 0
	MINOR       = 5
	PATCH       = 0
	ERR_COLOR   = "\033[1;38;2;255;17;0m"
	INFO_COLOR  = "\033[1;38;2;50;237;215m"
	OK_COLOR    = "\033[1;38;2;100;207;112m"
	TITLE_COLOR = "\033[1;38;2;138;120;255m"
)

var (
	VERSION_STRING = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, PATCH)
	encFlag        = flag.Bool("e", false, "Encrypt the provided file.")
	decFlag        = flag.Bool("d", false, "Decrypt the provided file.")
	versionFlag    = flag.Bool("v", false, "Version of this software ("+VERSION_STRING+")")
	fileFlag       = flag.String("f", "", "File name where input is read from")
	outFlag        = flag.String("o", "", "File name where output is written to.")
	seedFlag       = flag.String("s", "", "File name containing 128-bit seed and 96-bit nonce. Must be a binary file.")
	nonceFlag      = flag.String("n", "", "File name containing 128-bit seed. Must be a binary file.")
	txtFlag        = flag.Bool("t", false, "Save decrypted output as a text file")
)

func delayedEnd() {
	j := 0
	for j < 3 {
		fmt.Print("\033[1D.")
		time.Sleep(time.Duration(250) * time.Millisecond)
		fmt.Print("\033[1D ")
		time.Sleep(time.Duration(150) * time.Millisecond)
		j += 1
	}
	fmt.Print("\033[1D.")
}

func delayedPrint(text string, style string, delay_time float32, delay_end bool) {
	delay := 0
	for delay < 0 {
		delay = int(delay_time)
		delay_time *= 10.0
	}

	text_length := len(text)
	i := 0
	period, _ := utf8.DecodeRuneInString(".")
	fmt.Print(style)
	for i < text_length {
		time.Sleep(time.Duration(delay_time) * time.Millisecond)
		runeValue, w := utf8.DecodeRuneInString(text[i:])
		fmt.Printf("%c", runeValue)
		if delay_end && runeValue == period && i != text_length-1 {
			delayedEnd()
		}
		i += w
	}
	fmt.Print("\033[m\n")
}

func inputParams(seed []uint16, nonce []uint32) {
	bytebuf := make([]byte, 28)
	if *seedFlag != "" {
		file, err := os.ReadFile(*seedFlag)
		if err != nil {
			gameOver(err)
		}
		rbuf := bytes.NewReader(file[0:16])
		binary.Read(rbuf, binary.LittleEndian, &seed)
	} else {
		rand.Read(bytebuf[0:16])
		rbuf := bytes.NewReader(bytebuf[0:16])
		binary.Read(rbuf, binary.LittleEndian, &seed)
	}

	if *nonceFlag != "" {
		file, err := os.ReadFile(*nonceFlag)
		if err != nil {
			gameOver(err)
		}
		rbuf := bytes.NewReader(file[16:28])
		binary.Read(rbuf, binary.LittleEndian, &nonce)

	} else {
		rand.Read(bytebuf[16:28])
		rbuf := bytes.NewReader(bytebuf[16:28])
		binary.Read(rbuf, binary.LittleEndian, &nonce)
	}
}

func taskMaster(file []byte, operation string) {
	seed := make([]uint16, 8)
	nonce := make([]uint32, 3)

	inputParams(seed, nonce)

	switch operation {
	case "ENC":
		Lilith(file, seed, nonce)
	case "DEC":
		Lilith(file, seed, nonce)
	default:
		delayedPrint("❌ Missing or invalid operation. Exiting.", ERR_COLOR, 27, true)
		os.Exit(1)
	}
}

func gameOver(err error) {
	panic(string(ERR_COLOR) + err.Error() + "\033[m")
}

func ArgParse(argc int, args []string) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		gameOver(err)
	}

	width_string := strings.Split(string(out), " ")[1]

	width, _ := strconv.Atoi(strings.TrimSpace(width_string))

	width = (width / 2) - 4

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\033[%dC[OPTIONS]\n", width)
		flag.PrintDefaults()
	}

	if argc == 1 {
		delayedPrint("No file provided. Exiting.", ERR_COLOR+"❌ ", 20, true)
		os.Exit(0)
	}
	flag.Parse()

	if *versionFlag {
		fmt.Println(INFO_COLOR + "ℹ️ " + VERSION_STRING)
		os.Exit(0)
	}

	file, err := os.ReadFile(*fileFlag)
	if err != nil {
		gameOver(err)
	}

	taskMaster(file, *opFlag)
}
