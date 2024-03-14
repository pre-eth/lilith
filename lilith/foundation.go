package lilith

import "encoding/binary"

const (
	U32_MAX = 1 << 32

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

func bytesToU32(src []byte) []uint32 {
	out := make([]uint32, len(src)/4)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(src[i*4:])
	}
	return out
}

// For computing the indices needed by the combiner for dynamic folding
func dynamicIdx(arr []uint16, idx uint, mod uint16) uint16 {
	return (arr[idx] + arr[idx+1]) & mod
}

// Bitwise rotate left a 32-bit unsigned integer
func rotate(x uint32, shift byte) uint32 {
	return (x << shift) | ((x >> (32 - shift)) & (1<<shift - 1))
}

/*
Courtesy of: https://graphics.stanford.edu/~seander/bithacks.html#SwappingValuesXOR

This is an old trick to exchange the values of the variables a and b without using
extra space for a temporary variable.
*/
func swap(arr []byte, a uint, b uint) {
	arr[b] ^= arr[a] ^ arr[b]
	arr[a] ^= arr[b]
}

/*
The state and counter variables are initialized from the sub keys as follows
For original piecewise function, see img/INIT_PIECEWISE.png

NOTE: param <arr> is guaranteed by caller to be at least of size 8
*/
func initPiecewise(arr []uint16, h byte, i byte, j byte) uint32 {
	num := uint32(arr[(i+j)&7]) << 16
	num |= uint32(arr[(h+j)&7])
	return num
}

// phi() determines the value of the counter carry bit used in ctr_system()
// For original function that gives counter carry bit, see img/COUNTER_CARRY.png
func phi(ctr uint32, a uint32, cc *byte) uint32 {
	var sum uint64 = uint64(ctr + a + uint32(*cc))
	if sum >= U32_MAX {
		*cc = 1
	} else {
		*cc = 0
	}
	return uint32(*cc)
}

// The counter dynamics defined in Section 2.5, which you can also find in img/COUNTER_SYSTEM.png
func ctrSystem(ctr []uint32, _phi *byte) {
	ctr[0] = (ctr[0] + A0 + phi(ctr[0], A0, _phi)) & (U32_MAX - 1)
	ctr[1] = (ctr[1] + A1 + phi(ctr[1], A1, _phi)) & (U32_MAX - 1)
	ctr[2] = (ctr[2] + A2 + phi(ctr[2], A2, _phi)) & (U32_MAX - 1)
	ctr[3] = (ctr[3] + A3 + phi(ctr[3], A3, _phi)) & (U32_MAX - 1)
	ctr[4] = (ctr[4] + A4 + phi(ctr[4], A4, _phi)) & (U32_MAX - 1)
	ctr[5] = (ctr[5] + A5 + phi(ctr[5], A5, _phi)) & (U32_MAX - 1)
	ctr[6] = (ctr[6] + A6 + phi(ctr[6], A6, _phi)) & (U32_MAX - 1)
	ctr[7] = (ctr[7] + A7 + phi(ctr[7], A7, _phi)) & (U32_MAX - 1)
}

// Represents ((x + j)^2 XOR (((x + j)^2) >> 32)) mod 2^32
func gFunction(state []uint32, ctr []uint32, gfn []uint32) {
	var i byte = 0
	var tmp uint64
	for i < 8 {
		tmp = uint64(state[i]) + uint64(ctr[i])
		tmp *= tmp
		gfn[i] = uint32((tmp ^ (tmp >> 32)) & (U32_MAX - 1))
	}
}

func sboxMix(nonce uint32) byte {
	return byte(nonce&0xFF) ^ byte(nonce>>8) ^ byte(nonce>>16)
}

func srboxLookup(i uint16, sbox []byte, skey []byte) byte {
	idx := byte(i)
	return sbox[sbox[sbox[sbox[idx^skey[0]]^skey[1]]^skey[2]]^skey[3]]
}
