package lilith

import (
	"encoding/binary"
	"fmt"
)

type Lilith struct {
	// 	128-bit seed
	seed [8]uint32

	// 	96-bit nonce
	nonce [3]uint32

	// 	Counters for Rabbit PRNG
	ctr [8]uint32

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
	sbox [256]byte

	//	Current keystream
	key [8]uint32

	// 	The counter carry bit, phi, which needs to be stored between iterations.
	// 	This counter carry bit is initialized to zero:
	phi byte
}

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
