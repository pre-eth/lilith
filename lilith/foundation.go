package lilith

import "fmt"

// Following constants are used in the counter system to obtain the next counter
// for use with the next_state()
const(
	A0 = 0x4D34D34D
	A1 = 0xD34D34D3
	A2 = 0x34D34D34
	A3 = 0x4D34D34D
	A4 = 0xD34D34D3
	A5 = 0x34D34D34
	A6 = 0x4D34D34D
	A7 = 0xD34D34D3
)

func TestA(name string) {
	fmt.Printf("Hello %s", name)
}

