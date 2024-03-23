package lilith

import (
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
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
	Spinner    = "⣾⣽⣻⢿⡿⣟⣯⣷"
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
	jpgFlag               = flag.Bool("j", false, "Save decrypted output as a JPEG")
	pngFlag               = flag.Bool("p", false, "Save decrypted output as a PNG")
	quietFlag             = flag.Bool("q", false, "Quick mode - reduce output FX")
	textDelay     float32 = 20.0
	periodDelay           = true
	unit                  = "B"
	unitSize              = 1
	totalSize             = 0
)

// CLI spinner while performing operation
func spinner(written int) {
	idx := ((written >> 4) & 7) * 3
	r, _ := utf8.DecodeRuneInString(Spinner[idx:])
	fmt.Printf("%s%c\033[m\033[1D", InfoColor, r)
	if written > totalSize {
		written -= (written - totalSize)
	}

	progress := fmt.Sprintf(" %d/%d%s", written/unitSize, totalSize/unitSize, unit)
	fmt.Printf("\033[?25l\033[m %s\033[%dD", progress, len(progress)+1)
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

func saveDecrypted(outName string, data []byte) {
	data = data[:totalSize]
	if *txtFlag {
		//	If text flag set, interpret as text file
		fo, _ := os.Create(outName + ".txt")
		fmt.Println("\n" + string(data) + "\n")
		fo.WriteString(string(data))
		fo.Close()
	} else {
		if *jpgFlag {
			fo, _ := os.Create(outName + ".jpg")
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				fmt.Println(len(data))
				log.Fatalln(err)
			}
			var opts jpeg.Options
			opts.Quality = 100
			jpeg.Encode(fo, img, &opts)
			fo.Close()
		} else if *pngFlag {
			fo, _ := os.Create(outName + ".png")
			img, _, _ := image.Decode(bytes.NewReader(data))

			png.Encode(fo, img)
			fo.Close()
		} else {
			//	Otherwise, interpret as binary
			fo, _ := os.Create(outName)
			fo.Write(data)
			fo.Close()
		}
	}
}

func getSeed(outName string, seed *[16]byte) {
	if *seedFlag != "" {
		file, err := os.ReadFile(*seedFlag)
		if err != nil {
			gameOver(err)
		}
		copy(seed[0:], file[0:])
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
				fmt.Println(totalSize)

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
			copy(seed[0:], file[0:16])
		}
	}
}

func getNonce(outName string, nonce *[12]byte) {
	if *nonceFlag != "" {
		file, err := os.ReadFile(*nonceFlag)
		if err != nil {
			gameOver(err)
		}
		copy(nonce[0:], file[0:])
	} else {
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
}

func taskMaster(filename string) {
	seed := [16]byte{}
	nonce := [12]byte{}

	// 	Get output file name
	outName := *outFlag
	if outName == "" {
		outName = "_output"
	}

	//	Check if bad file and set output unit for CLI
	inputData, err := os.ReadFile(filename)
	if err != nil {
		gameOver(err)
	}

	totalSize = len(inputData)

	if totalSize > 2000 {
		unit = "KB"
		unitSize = 1000
	}
	if totalSize > 2000000 {
		unit = "MB"
		unitSize = 1000000
	}
	if totalSize > 2000000000 {
		unit = "GB"
		unitSize = 100000000
	}

	getSeed(outName, &seed)
	getNonce(outName, &nonce)

	//	VALIDATION OVER
	// 	Initialize and perform requested operation

	lilith := Lilith{}
	lilith.Init(&seed, &nonce, *decFlag, inputData[0])

	if *encFlag {
		delayedPrint("LILITH "+VersionString+" - ENCRYPT ", TitleColor, textDelay, false)

		ciphertext := lilith.Encrypt(inputData)

		fmt.Print("\033[1D\033[m\n\n")
		delayedPrint("Completed encryption.\n", OkColor, textDelay, periodDelay)

		//	Save encrypted output
		fo, _ := os.Create(outName)
		fo.Write(ciphertext)
		fo.Close()

	} else {
		delayedPrint("LILITH "+VersionString+" - DECRYPT ", TitleColor, textDelay, false)

		plaintext := lilith.Decrypt(inputData)

		fmt.Print("\033[1D\033[m\n\n")
		delayedPrint("Completed decryption.\n", OkColor, textDelay, periodDelay)

		saveDecrypted(outName, plaintext)
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
