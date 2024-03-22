package lilith

const (
	U32_MAX = 1 << 32

	// 	Following constants are used in the counter system to obtain the next counter
	// 	for use with the next_state()

	A0 = 0x4D34D34D
	A1 = 0xD34D34D3
	A2 = 0x34D34D34
	A3 = 0x4D34D34D
	A4 = 0xD34D34D3
	A5 = 0x34D34D34
	A6 = 0x4D34D34D
	A7 = 0xD34D34D3
)

func bytesToU32(out *[8]uint32, src []byte) {
	i := 0
	for ; i < 16; i += 2 {
		out[i>>1] = uint32(src[i+1])<<8 | uint32(src[i])
	}
}

// For computing the indices needed by the combiner for dynamic folding
func dynamicIdx(arr *[8]uint32, idx uint, mod uint32) int {
	return int((arr[idx] + arr[idx+1]) & mod)
}

// Bitwise rotate left a 32-bit unsigned integer
func rotate(x uint32, shift byte) uint32 {
	return (x << shift) | ((x >> (32 - shift)) & (1<<shift - 1))
}

func swap(arr []byte, a uint, b uint) {
	tmp := arr[a]
	arr[a] = arr[b]
	arr[b] = tmp
}

func initPiecewise(arr *[8]uint32, h byte, i byte, j byte) uint32 {
	/*
		The state and counter variables are initialized from the sub keys as follows
		For original piecewise function, see img/INIT_PIECEWISE.png

		NOTE: param <arr> is guaranteed by caller to be at least of size 8
	*/
	num := arr[(i+j)&7] << 16
	num |= arr[(h+j)&7]
	return num
}

func phi(ctr uint32, a uint32, cc *byte) uint32 {
	// 	phi() determines the value of the counter carry bit used in ctr_system()
	// 	For original function that gives counter carry bit, see img/COUNTER_CARRY.png
	var sum uint64 = uint64(ctr + a + uint32(*cc))
	if sum >= U32_MAX {
		*cc = 1
	} else {
		*cc = 0
	}
	return uint32(*cc)
}

func ctrSystem(ctr *[8]uint32, _phi *byte) {
	// 	The counter dynamics defined in Section 2.5, which you can also find in img/COUNTER_SYSTEM.png
	ctr[0] = (ctr[0] + A0 + phi(ctr[0], A0, _phi)) & (U32_MAX - 1)
	ctr[1] = (ctr[1] + A1 + phi(ctr[1], A1, _phi)) & (U32_MAX - 1)
	ctr[2] = (ctr[2] + A2 + phi(ctr[2], A2, _phi)) & (U32_MAX - 1)
	ctr[3] = (ctr[3] + A3 + phi(ctr[3], A3, _phi)) & (U32_MAX - 1)
	ctr[4] = (ctr[4] + A4 + phi(ctr[4], A4, _phi)) & (U32_MAX - 1)
	ctr[5] = (ctr[5] + A5 + phi(ctr[5], A5, _phi)) & (U32_MAX - 1)
	ctr[6] = (ctr[6] + A6 + phi(ctr[6], A6, _phi)) & (U32_MAX - 1)
	ctr[7] = (ctr[7] + A7 + phi(ctr[7], A7, _phi)) & (U32_MAX - 1)
}

func gFunction(state *[8]uint32, ctr *[8]uint32, gfn *[8]uint32) {
	// 	Represents ((x + j)^2 XOR (((x + j)^2) >> 32)) mod 2^32
	var i byte = 0
	var tmp uint64
	for i < 8 {
		tmp = uint64(state[i] + ctr[i])
		tmp *= tmp
		gfn[i] = uint32((tmp ^ (tmp >> 32)) & (U32_MAX - 1))
		i += 1
	}
}

func sboxMix(nonce uint32) byte {
	return byte(nonce) ^ byte(nonce>>8) ^ byte(nonce>>16)
}

func srboxLookup(i byte, sbox *[256]byte, skey *[4]byte) byte {
	return sbox[sbox[sbox[sbox[i^skey[0]]^skey[1]]^skey[2]]^skey[3]]
}

func setup(seed *[8]uint32, state *[8]uint32, ctr *[8]uint32) {
	/*
		The algorithm is initialized by extending the 128-bit key into both the eight 64-bit
		state variables and the eight 64-bit counters, yielding a one-to-one correspondence
		between the key and the initial state variables

		The key K[127…0] is divided into 8 sub keys:
		K0 = k[15...0], K1 = k[31...16], … K7 = k[127...112]
	*/

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

func nextState(state *[8]uint32, ctr *[8]uint32, phi *byte) {
	// 	For illustration of what this code is representing, refer to
	// 	img/NEXT_STATE.png which shows the PRNG's coupled system

	ctrSystem(ctr, phi)

	gfn := [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
	gFunction(state, ctr, &gfn)

	state[0] = gfn[0] + rotate(gfn[7], 16) + rotate(gfn[6], 16)
	state[1] = gfn[1] + rotate(gfn[0], 8) + gfn[7]
	state[2] = gfn[2] + rotate(gfn[1], 16) + rotate(gfn[0], 16)
	state[3] = gfn[3] + rotate(gfn[2], 8) + gfn[1]
	state[4] = gfn[4] + rotate(gfn[3], 16) + rotate(gfn[2], 16)
	state[5] = gfn[5] + rotate(gfn[4], 8) + gfn[3]
	state[6] = gfn[6] + rotate(gfn[5], 16) + rotate(gfn[4], 16)
	state[7] = gfn[7] + rotate(gfn[6], 8) + gfn[5]
}

func extractKeystream(key *[8]uint32, state *[8]uint32) {
	/*
		Implementation of the Extraction Scheme in Section 2.6

		of the Rabbit cipher

		At the end of the third iteration of next_state, we extract
		128-bits of output from the state variables in the following
		manner to be used as our 16 byte keystream for next iteration

		For a clearer depiction, see img/EXTRACT_KS.png
	*/

	key[0] = (state[0] & 0xFFFF) ^ ((state[5] >> 16) & 0xFFFF)
	key[1] = ((state[0] >> 16) & 0xFFFF) ^ (state[3] & 0xFFFF)
	key[2] = (state[2] & 0xFFFF) ^ ((state[7] >> 16) & 0xFFFF)
	key[3] = ((state[2] >> 16) & 0xFFFF) ^ (state[5] & 0xFFFF)
	key[4] = (state[4] & 0xFFFF) ^ ((state[1] >> 16) & 0xFFFF)
	key[5] = ((state[4] >> 16) & 0xFFFF) ^ (state[7] & 0xFFFF)
	key[6] = (state[6] & 0xFFFF) ^ ((state[3] >> 16) & 0xFFFF)
	key[7] = ((state[6] >> 16) & 0xFFFF) ^ (state[1] & 0xFFFF)
}

func generateC0(ctext *[16]byte, key *[8]uint32, sbox *[256]byte, operation bool) {
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
	idx := 4 + int(key[7]&15)
	start_ptext := P0[idx:]
	start_ptext = append(start_ptext, P0[:16-(21-idx)]...)
	combiner(key, start_ptext, sbox)
	copy(ctext[:], start_ptext[:16])
}

func generateSbox(sbox *[256]byte, nonce *[3]uint32, operation bool) [256]byte {
	key_bytes := [4]byte{sboxMix(nonce[0]), sboxMix(nonce[1]), sboxMix(nonce[2])}
	key_bytes[3] = byte(nonce[0]>>24) ^ byte(nonce[1]>>24) ^ byte(nonce[2]>>24)

	new_sbox := [256]byte{}

	var i uint16 = 0
	for i < 256 {
		new_sbox[i] = srboxLookup(byte(i), sbox, &key_bytes)
		i += 1
	}

	//	Need to create INVERSE S-box for decryption
	if operation {
		sbox_idx := [256]byte{}

		i = 0
		for i < 256 {
			old := new_sbox[i]
			row := old & 15
			col := (old >> 4) & 15
			idx := (col << 4) + row
			idx = new_sbox[idx]
			row = idx & 15
			col = (idx >> 4) & 15
			idx = (col << 4) + row
			sbox_idx[idx] = old
			i += 1
		}

		new_sbox = sbox_idx
	}

	return new_sbox
}
