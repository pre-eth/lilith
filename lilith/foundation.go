package lilith

const(
	U32_MAX = 1<<32

	/*
		Text 1 (𝑃0) is used only once in the algorithm. It is 16-bytes text utilized 
		as a starting (virtual) plaintext which is input to the encryption/decryption 
		function with KS0 to produce C0, the starting (virtual) ciphertext which is used 
		as feedback into the PRNG to finalize the modified key setup and produce the 
		first keystream used for encryption, KS1.

		The P0 chosen for this implementation is the word "Lilith" written in Hebrew, and
		is actually 21 bytes instead of 16. However, the algorithm is still satisfied by
		taking KS0 mod 16, using that as the starting point. If there are less than 16 bytes
		available for extraction from that starting point ((KS0 mod 16) > 5), then we collect
		all bytes from starting position and then wrap back around to fill the remaining bytes

		NOTE: KS0, P0, and C0 is known to both the encipherer and decipherer. So using this text
		was just an arbitrary choice :P
	*/
	P0 = "נידה ,לילית"
	P0_LEN = 21

// Following constants are used in the counter system to obtain the next counter
// for use with the next_state()
	A0 = 0x4D34D34D
	A1 = 0xD34D34D3
	A2 = 0x34D34D34
	A3 = 0x4D34D34D
	A4 = 0xD34D34D3
	A5 = 0x34D34D34
	A6 = 0x4D34D34D
	A7 = 0xD34D34D3
)

// For computing the indices needed by the combiner for dynamic folding
func DynamicIdx(arr []uint, idx uint, mod uint) uint {
	return (arr[idx] + arr[idx + 1]) & mod
}

// Bitwise rotate left a 32-bit unsigned integer
func Rotate(x uint32, shift uint8) uint32 {
	return (x << shift) | ((x >> (32 - shift)) & (1<<shift - 1))
}

// Courtesy of: https://graphics.stanford.edu/~seander/bithacks.html#SwappingValuesXOR
// 
// This is an old trick to exchange the values of the variables a and b without using 
// extra space for a temporary variable.
func Swap(arr []uint, a uint, b uint) {
	arr[b] ^= arr[a] ^ arr[b]
	arr[a] ^= arr[b]
}

// The state and counter variables are initialized from the sub keys as follows:
// For original piecewise function, see img/INIT_PIECEWISE.png
func InitPiecewise(arr []uint16, h uint8, i uint8, j uint8) uint32 {
	num := uint32(arr[(i + j) & 7]) << 16
	num |= uint32(arr[(h + j) & 7])
	return num
}

