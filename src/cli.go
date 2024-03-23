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
	MAJOR      = 0
	MINOR      = 5
	PATCH      = 0
	ErrColor   = "\033[1;38;2;255;17;0m"
	InfoColor  = "\033[1;38;2;50;237;215m"
	OkColor    = "\033[1;38;2;100;207;112m"
	TitleColor = "\033[1;38;2;165;97;255m"
)

var (
	VersionString         = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, PATCH)
	encFlag               = flag.Bool("e", false, "Encrypt the provided file.")
	decFlag               = flag.Bool("d", false, "Decrypt the provided file.")
	versionFlag           = flag.Bool("v", false, "Version of this software ("+VersionString+")")
	fileFlag              = flag.String("f", "", "File name where input is read from")
	outFlag               = flag.String("o", "", "File name where output is written to.")
	seedFlag              = flag.String("s", "", "File name containing 128-bit seed. Must be a binary file.")
	nonceFlag             = flag.String("n", "", "File name containing 96-bit nonce. Must be a binary file.")
	txtFlag               = flag.Bool("t", false, "Save decrypted output as a text file")
	quietFlag             = flag.Bool("q", false, "Less verbose and interactive output")
	textDelay     float32 = 20.0
	periodDelay           = true
)

// CLI spinner while performing operation
func spinner() {
	spinnerString := "⣾⣽⣻⢿⡿⣟⣯⣷"
	fmt.Print(InfoColor)
	for _, s := range spinnerString {
		fmt.Printf("%c", s)
		time.Sleep(time.Duration(20) * time.Millisecond)
		fmt.Print("\033[1D")
	}
	fmt.Print("\033[m")
}

// Period blink animation
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
	fmt.Print("\033[m")
}

func taskMaster(filename string) {
	seed := [16]byte{}
	nonce := [12]byte{}

	// 	Get output file name
	outName := *outFlag
	if outName == "" {
		outName = "_output"
	}

	if *seedFlag != "" {
		file, err := os.ReadFile(*seedFlag)
		if err != nil {
			gameOver(err)
		}
		copy(seed[:], file[:])
	} else {
		possible := outName + ".lseed"

		if *decFlag {
			possible = *fileFlag + ".lseed"
		}

		// 	Check if a .lseed file exists with same name as this provided file
		//	Lets user avoid specifying init params that retain the default names
		if _, err := os.Stat(possible); errors.Is(err, os.ErrNotExist) {
			if !*decFlag {
				rand.Read(seed[0:])
				fo, _ := os.Create(possible)
				fo.Write(seed[0:])
				fo.Close()
			} else {
				delayedPrint("Missing key and nonce parameters for decryption.\n", ErrColor, textDelay, periodDelay)
				os.Exit(1)
			}
		} else {
			file, err := os.ReadFile(possible)
			if err != nil {
				gameOver(err)
			}
			copy(seed[0:], file[0:])
		}
	}

	if *nonceFlag != "" {
		file, err := os.ReadFile(*nonceFlag)
		if err != nil {
			gameOver(err)
		}
		copy(nonce[:], file[:])
	} else {
		//	Equivalent logic for above but for nonces

		possible := outName + ".lnonce"

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
				delayedPrint("Missing key and nonce parameters for decryption.\n", ErrColor, textDelay, periodDelay)
				os.Exit(1)
			}
		} else {
			file, err := os.ReadFile(possible)
			if err != nil {
				gameOver(err)
			}
			copy(nonce[0:], file[0:])
		}
	}

	//	Bad file?
	file, err := os.ReadFile(filename)
	if err != nil {
		gameOver(err)
	}

	//	VALIDATION OVER
	// 	Initialize and perform requested operation

	lilith := Lilith{}
	lilith.Init(&seed, &nonce, *decFlag, file[0])

	if *encFlag {
		ciphertext := lilith.Encrypt(file)

		//	Save encrypted output
		fo, _ := os.Create(outName)
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
			fo, _ := os.Create(outName)
			fmt.Println("\n" + string(plaintext[:last+1]) + "\n")
			fo.WriteString(string(plaintext[:last+1]))
			fo.Close()
		} else {
			//	Otherwise, interpret as binary
			fo, _ := os.Create(outName)
			fo.Write(plaintext)
			fo.Close()
		}
	}

	delayedPrint("Output written to "+outName+".\n\n", InfoColor, textDelay, periodDelay)
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

	// Quiet mode?
	if *quietFlag {
		textDelay = 0.0
		periodDelay = false
	}

	//	Print version?
	if *versionFlag {
		fmt.Println(InfoColor + VersionString)
		os.Exit(0)
	}

	//	Validate operation
	if *encFlag && *decFlag {
		delayedPrint("Only operation may be specified at a time.\n", ErrColor, textDelay, periodDelay)
		os.Exit(1)
	} else if !*encFlag && !*decFlag {
		delayedPrint("Missing or invalid operation.\n", ErrColor, textDelay, periodDelay)
		os.Exit(1)
	}

	//	See if input file name provided
	if *fileFlag == "" {
		//	No file!
		delayedPrint("No file provided.\n", ErrColor, textDelay, periodDelay)
		os.Exit(1)
	}

	// Perform the requested operation
	taskMaster(*fileFlag)
}

func gameOver(err error) {
	panic(string(ErrColor) + err.Error() + "\033[m\n")
}
