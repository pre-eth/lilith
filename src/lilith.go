package lilith

import (
	"fmt"
)

/*
The algorithm is initialized by extending the 128-bit key into both the eight 64-bit
state variables and the eight 64-bit counters, yielding a one-to-one correspondence
between the key and the initial state variables

The key K[127…0] is divided into 8 sub keys:
K0 = k[15...0], K1 = k[31...16], … K7 = k[127...112]
*/
func setup(seed []uint16, state []uint32, ctr []uint32) {
	// STATE
	state[0] = initPiecewise(seed, 0, 1, 0)
	state[1] = initPiecewise(seed, 1, 5, 4)
	state[2] = initPiecewise(seed, 2, 1, 0)
	state[3] = initPiecewise(seed, 3, 5, 4)
	state[4] = initPiecewise(seed, 4, 1, 0)
	state[5] = initPiecewise(seed, 5, 5, 4)
	state[6] = initPiecewise(seed, 6, 1, 0)
	state[7] = initPiecewise(seed, 7, 5, 4)

	// CTR
	ctr[0] = initPiecewise(seed, 0, 4, 5)
	ctr[1] = initPiecewise(seed, 1, 0, 1)
	ctr[2] = initPiecewise(seed, 2, 4, 5)
	ctr[3] = initPiecewise(seed, 3, 0, 1)
	ctr[4] = initPiecewise(seed, 4, 4, 5)
	ctr[5] = initPiecewise(seed, 5, 0, 1)
	ctr[6] = initPiecewise(seed, 6, 4, 5)
	ctr[7] = initPiecewise(seed, 7, 0, 1)
}

// For illustration of what this code is representing, refer to
// img/NEXT_STATE.png which shows the PRNG's coupled system
func nextState(state []uint32, ctr []uint32, phi *byte) {
	ctrSystem(ctr, phi)

	gfn := []uint32{0, 0, 0, 0, 0, 0, 0, 0}
	gFunction(state, ctr, gfn)

	state[0] = gfn[0] + rotate(gfn[7], 16) + rotate(gfn[6], 16)
	state[1] = gfn[1] + rotate(gfn[0], 8) + gfn[7]
	state[2] = gfn[2] + rotate(gfn[1], 16) + rotate(gfn[0], 16)
	state[3] = gfn[3] + rotate(gfn[2], 8) + gfn[1]
	state[4] = gfn[4] + rotate(gfn[3], 16) + rotate(gfn[2], 16)
	state[5] = gfn[5] + rotate(gfn[4], 8) + gfn[3]
	state[6] = gfn[6] + rotate(gfn[5], 16) + rotate(gfn[4], 16)
	state[7] = gfn[7] + rotate(gfn[6], 8) + gfn[5]
}

/*
Implementation of the Extraction Scheme in Section 2.6
of the Rabbit cipher

At the end of the third iteration of next_state, we extract
128-bits of output from the state variables in the following
manner to be used as our 16 byte keystream for next iteration

For a clearer depiction, see img/EXTRACT_KS.png
*/
func extractKeystream(key []uint16, state []uint32) {
	key[0] = uint16((state[0] & 0xFFFF) ^ ((state[5] >> 16) & 0xFFFF))
	key[1] = uint16(((state[0] >> 16) & 0xFFFF) ^ (state[3] & 0xFFFF))
	key[2] = uint16((state[2] & 0xFFFF) ^ ((state[7] >> 16) & 0xFFFF))
	key[3] = uint16(((state[2] >> 16) & 0xFFFF) ^ (state[5] & 0xFFFF))
	key[4] = uint16((state[4] & 0xFFFF) ^ ((state[1] >> 16) & 0xFFFF))
	key[5] = uint16(((state[4] >> 16) & 0xFFFF) ^ (state[7] & 0xFFFF))
	key[6] = uint16((state[6] & 0xFFFF) ^ ((state[3] >> 16) & 0xFFFF))
	key[7] = uint16(((state[6] >> 16) & 0xFFFF) ^ (state[1] & 0xFFFF))
}

func generateC0(ctext []byte, key []uint16, sbox []byte) {
	/*
		Text 1 (𝑃0) is used only once in the algorithm. It is 16-bytes text utilized
		as a starting (virtual) plaintext which is input to the encryption/decryption
		function with KS0 to produce C0, the starting (virtual) ciphertext which is used
		as feedback into the PRNG to finalize the modified key setup and produce the
		first keystream used for encryption, KS1.

		The P0 chosen for this implementation is "נידה ,לילית" - the word "Lilith" written
		in Hebrew, and is actually 21 bytes instead of 16. However, the algorithm is still
		satisfied by taking KS0 mod 16, using that as the starting point. If there's less than
		16 bytes available from that starting point ((KS0 mod 16) > 5), then we collect all
		bytes from starting position and then wrap back around to fill the remaining bytes

		NOTE: KS0, P0, and C0 is known to both the encipherer and decipherer. So using this
		text was just an arbitrary choice :P
	*/
	P0 := [21]byte{0xD7, 0xA0, 0xD7, 0x99, 0xD7, 0x93, 0xD7, 0x94, 0x20, 0x2C, 0xD7, 0x9C, 0xD7, 0x99, 0xD7, 0x9C, 0xD7, 0x99, 0xD7, 0xAA}

	// Write 16 bytes from P0 macro to use as initial plain_text
	// Start at 5 to never write more than 16 bytes in one go
	var idx byte = 5 + byte(key[7]&15)
	start_ptext := P0[idx:]
	copy(start_ptext, P0[:16-(21-idx)])

	combiner(key, ctext, sbox)
}

func generateSbox(sbox []byte, state []uint32) {
	key_bytes := make([]byte, 4)

	key_bytes[0] = sboxMix(state[0])
	key_bytes[1] = sboxMix(state[1])
	key_bytes[2] = sboxMix(state[2])
	key_bytes[3] = byte(state[0]>>24) ^ byte(state[1]>>24) ^ byte(state[2]>>24)

	var i uint16 = 0
	for i < 256 {
		sbox[i+0] = srboxLookup(i+0, sbox, key_bytes)
		sbox[i+1] = srboxLookup(i+1, sbox, key_bytes)
		sbox[i+2] = srboxLookup(i+2, sbox, key_bytes)
		sbox[i+3] = srboxLookup(i+3, sbox, key_bytes)
		sbox[i+4] = srboxLookup(i+4, sbox, key_bytes)
		sbox[i+5] = srboxLookup(i+5, sbox, key_bytes)
		sbox[i+6] = srboxLookup(i+6, sbox, key_bytes)
		sbox[i+7] = srboxLookup(i+7, sbox, key_bytes)
		i += 8
	}
}

func Lilith(ciphertext []byte, seed []uint16, nonce []uint32) []uint16 {
	fmt.Println("LILITH")

	// Initial state and counter setup
	state := make([]uint32, 8)
	ctr := make([]uint32, 8)
	setup(seed, state, ctr)

	// The counter carry bit, phi, which needs to be stored between iterations.
	// This counter carry bit is initialized to zero:
	var phi byte = 0

	// Paper says to call nextState() 3x to finish initialization
	nextState(state, ctr, &phi)
	nextState(state, ctr, &phi)
	nextState(state, ctr, &phi)

	// Get KS0
	extractKeystream(seed, state)

	/*
		Fixed one dimensional representation of the Rijndael S-box shown in img/SRBOX.png
		Rijndael has been designed to be resistant to known cryptanalytic attacks

		This is used in generating a private S-box for the encryption scheme. The choice of
		the S-box provides a much stronger diffusion. Each output bit depends on each input
		bit of the S-box key. In the decryption process, the inverse Sub Bytes transformation
		is applied. The same transformation is implemented but with use of the inverse
		substitution box.

		Since S-box is private, an attacker cannot know the input to the S-box. The 96-bit
		nonce is used in the shuffle process to initialize the S-Bwox
	*/
	SBOX := []byte{
		0x63, 0x7C, 0x77, 0x7B, 0xF2, 0x6B, 0x6F, 0xC5, 0x30, 0x01, 0x67, 0x2B, 0xFE, 0xD7, 0xAB, 0x76,
		0xCA, 0x82, 0xC9, 0x7D, 0xFA, 0x59, 0x47, 0xF0, 0xAD, 0xD4, 0xA2, 0xAF, 0x9C, 0xA4, 0x72, 0xC0,
		0xB7, 0xFD, 0x93, 0x26, 0x36, 0x3F, 0xF7, 0xCC, 0x34, 0xA5, 0xE5, 0xF1, 0x71, 0xD8, 0x31, 0x15,
		0x04, 0xC7, 0x23, 0xC3, 0x18, 0x96, 0x05, 0x9A, 0x07, 0x12, 0x80, 0xE2, 0xEB, 0x27, 0xB2, 0x75,
		0x09, 0x83, 0x2C, 0x1A, 0x1B, 0x6E, 0x5A, 0xA0, 0x52, 0x3B, 0xD6, 0xB3, 0x29, 0xE3, 0x2F, 0x84,
		0x53, 0xD1, 0x00, 0xED, 0x20, 0xFC, 0xB1, 0x5B, 0x6A, 0xCB, 0xBE, 0x39, 0x4A, 0x4C, 0x58, 0xCF,
		0xD0, 0xEF, 0xAA, 0xFB, 0x43, 0x4D, 0x33, 0x85, 0x45, 0xF9, 0x02, 0x7F, 0x50, 0x3C, 0x9F, 0xA8,
		0x51, 0xA3, 0x40, 0x8F, 0x92, 0x9D, 0x38, 0xF5, 0xBC, 0xB6, 0xDA, 0x21, 0x10, 0xFF, 0xF3, 0xD2,
		0xCD, 0x0C, 0x13, 0xEC, 0x5F, 0x97, 0x44, 0x17, 0xCA, 0xA7, 0x7E, 0x3D, 0x64, 0x5D, 0x19, 0x73,
		0x60, 0x81, 0x4F, 0xDC, 0x22, 0x2A, 0x90, 0x88, 0x46, 0xEE, 0xB8, 0x14, 0xDE, 0x5E, 0x08, 0xD8,
		0xE0, 0x32, 0x3A, 0x0A, 0x49, 0x06, 0x24, 0x5C, 0xC2, 0xD3, 0xAC, 0x62, 0x91, 0x95, 0xEA, 0x79,
		0xE7, 0xC8, 0x37, 0x6D, 0x8D, 0xD5, 0x4E, 0xA9, 0x6C, 0x56, 0xF4, 0xEA, 0x65, 0x7A, 0xAE, 0x08,
		0xBA, 0x78, 0x25, 0x2E, 0x1C, 0xA6, 0xB4, 0xC6, 0xE8, 0xDD, 0x74, 0x1F, 0x4B, 0xBD, 0x8B, 0x8A,
		0x70, 0x3E, 0xB5, 0x66, 0x48, 0x03, 0xF6, 0x0E, 0x61, 0x35, 0x57, 0xB9, 0x86, 0xC1, 0x1D, 0x9E,
		0xE1, 0xF8, 0x98, 0x11, 0x69, 0xD9, 0x8E, 0x94, 0x9B, 0x1E, 0x87, 0xE9, 0xCE, 0x55, 0x28, 0xDF,
		0x8C, 0xA1, 0x89, 0x0D, 0xBF, 0xE6, 0x42, 0x68, 0x41, 0x99, 0x2D, 0x0F, 0xB0, 0x54, 0xBB, 0x16}

	generateSbox(SBOX, nonce)

	// Generate the first cipher text
	C0 := make([]byte, 16)
	generateC0(C0, seed, SBOX)

	// Generate and store KS1 (needed in the future by decipherer)
	decrypt_key := make([]uint16, 8)
	bytes_u32 := bytesToU32(C0)
	nextState(bytes_u32, ctr, &phi)
	extractKeystream(decrypt_key, bytes_u32)

	i := 0
	key_stream := make([]uint16, 8)
	copy(key_stream, decrypt_key)
	ctext_len := len(ciphertext)

	// Logic depends on bytes of 16 blocks, so process leftover blocks differently
	ctext_len -= ctext_len & 15

	for i < ctext_len {
		combiner(key_stream, ciphertext[i:], SBOX)
		nextState(bytes_u32, ctr, &phi)
		extractKeystream(key_stream, bytes_u32)
		i += 16
	}

	return decrypt_key
}
