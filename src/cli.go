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
	TITLE_COLOR = "\033[1;38;2;165;97;255m"
)

var (
	VERSION_STRING = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, PATCH)
	encFlag        = flag.Bool("e", false, "Encrypt the provided file.")
	decFlag        = flag.Bool("d", false, "Decrypt the provided file.")
	versionFlag    = flag.Bool("v", false, "Version of this software ("+VERSION_STRING+")")
	fileFlag       = flag.String("f", "", "File name where input is read from")
	outFlag        = flag.String("o", "", "File name where output is written to.")
	seedFlag       = flag.String("s", "", "File name containing 128-bit seed. Must be a binary file.")
	nonceFlag      = flag.String("n", "", "File name containing 96-bit nonce. Must be a binary file.")
	txtFlag        = flag.Bool("t", false, "Save decrypted output as a text file")
)

func delayedEnd() {
	j := 0
	for j < 2 {
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
		if delay_end && runeValue == period {
			delayedEnd()
		}
		i += w
	}
	fmt.Print("\033[m\n")
}

func taskMaster(filename string) {
	seed := [16]byte{}
	nonce := [12]byte{}

	// 	Get output file name
	out_name := *outFlag
	if out_name == "" {
		out_name = "_output"
	}

	if *seedFlag != "" {
		file, err := os.ReadFile(*seedFlag)
		if err != nil {
			gameOver(err)
		}
		copy(seed[:], file[:])
	} else {
		possible := out_name + ".lseed"

		if *decFlag {
			possible = *fileFlag + ".lseed"
		}

		// 	Check if a .lnonce file exists with same name as this provided file
		//	Lets user avoid specifying init params that retain the default names
		if _, err := os.Stat(possible); errors.Is(err, os.ErrNotExist) {
			if !*decFlag {
				rand.Read(seed[0:])
				fo, _ := os.Create(possible)
				fo.Write(seed[0:])
				fo.Close()
			} else {
				delayedPrint("Missing key and nonce parameters for decryption.", ERR_COLOR, 20, true)
				os.Exit(1)
			}
		} else {
			file, err := os.ReadFile(possible)
			if err != nil {
				gameOver(err)
			}
			copy(seed[0:], file[:])
		}
	}

	if *nonceFlag != "" {
		file, err := os.ReadFile(*nonceFlag)
		if err != nil {
			gameOver(err)
		}
		copy(nonce[:], file[:])
	} else {
		possible := out_name + ".lnonce"

		if *decFlag {
			possible = *fileFlag + ".lnonce"
		}

		if _, err := os.Stat(possible); errors.Is(err, os.ErrNotExist) {
			if !*decFlag {
				rand.Read(nonce[0:])
				fo, _ := os.Create(possible)
				fo.Write(nonce[0:])
				fo.Close()
			} else {
				delayedPrint("Missing key and nonce parameters for decryption.", ERR_COLOR, 20, true)
				os.Exit(1)
			}
		} else {
			file, err := os.ReadFile(possible)
			if err != nil {
				gameOver(err)
			}
			copy(nonce[0:], file[:])
		}
	}

	//	Bad file?
	file, err := os.ReadFile(filename)
	if err != nil {
		gameOver(err)
	}

	lilith := Lilith{}
	lilith.Init(&seed, &nonce, *decFlag, file[0])

	if *encFlag {
		ciphertext := lilith.Encrypt(file)

		//	Save encrypted output
		fo, _ := os.Create(out_name)
		fo.Write(ciphertext)
		fo.Close()
	} else {
		plaintext := lilith.Decrypt(file)

		last := len(plaintext) - 1
		for last > 0 {
			if plaintext[last] == 0 {
				last -= 1
				continue
			}
			break
		}

		if *txtFlag {
			//	If text flag set, interpret as text file
			fo, _ := os.Create(out_name + ".txt")
			fmt.Println(string(plaintext[:last+1]))
			fo.WriteString(string(plaintext[:last+1]))
			fo.Close()
		} else {
			//	Otherwise, interpret as binary
			fo, _ := os.Create(out_name)
			fo.Write(plaintext)
			fo.Close()
		}
	}
}

func ArgParse(argc int, args []string) {
	//	Get terminal size for centering the help header
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		gameOver(err)
	}

	//	Parse size to int
	width_string := strings.Split(string(out), " ")[1]
	width, _ := strconv.Atoi(strings.TrimSpace(width_string))
	width = (width / 2) - 4

	//	Prepare -h contents
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\033[%dC[OPTIONS]\n", width)
		flag.PrintDefaults()
	}

	flag.Parse()

	//	Print version
	if *versionFlag {
		fmt.Println(INFO_COLOR + VERSION_STRING)
		os.Exit(0)
	}

	//	Validate operation
	if *encFlag && *decFlag {
		delayedPrint("Only operation may be specified at a time.", ERR_COLOR, 20, true)
		os.Exit(1)
	} else if !*encFlag && !*decFlag {
		delayedPrint("Missing or invalid operation.", ERR_COLOR, 20, true)
		os.Exit(1)
	}

	//	See if input file name provided
	if *fileFlag == "" {
		//	No file!
		delayedPrint("No file provided.", ERR_COLOR, 20, true)
		os.Exit(1)
	}

	// Perform the requested operation
	taskMaster(*fileFlag)
}

func gameOver(err error) {
	panic(string(ERR_COLOR) + err.Error() + "\033[m")
}
