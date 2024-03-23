package lilith

import (
	"encoding/binary"
)

/*
	For the combiner algorithm, the length of the cipher key is 128 bits. Each
	round consists of four functions:
	  a) add round keystream                (img/ADD_ROUND.png)
	  b) byte substitution                  (img/BYTE_SUBSTITUION.png)
	  c) shift rows                         (img/SHIFT_ROWS.png)
	  d) dynamic folding transformations    (img/DYNAMIC_FOLD.png)

	  See img/COMBINER.png for llustrations of the full process

a. 	Add Round Keystream Transformation: The keystream generated from the

	PRNG is used byte by byte, from lowest to highest index, so there is no need
	for keystream array to be in a 2-dimentional form; just use them up and
	move on. The function Add Round Keystream uses 16 bytes of expanded key every
	time it is called. The operation of the inverse Add Round Keystream
	transformation is simply applied by performing the same forward transformation
	since Add Round Keystream is its own inverse.

	NOTE: The parameter ptext will be refered to as the "state" from now on
*/
func addRoundKs(key *[8]uint32, ptext []byte) {
	ptext[0] ^= byte(key[0])
	ptext[1] ^= byte(key[0] >> 8)
	ptext[2] ^= byte(key[1])
	ptext[3] ^= byte(key[1] >> 8)
	ptext[4] ^= byte(key[2])
	ptext[5] ^= byte(key[2] >> 8)
	ptext[6] ^= byte(key[3])
	ptext[7] ^= byte(key[3] >> 8)
	ptext[8] ^= byte(key[4])
	ptext[9] ^= byte(key[4] >> 8)
	ptext[10] ^= byte(key[5])
	ptext[11] ^= byte(key[5] >> 8)
	ptext[12] ^= byte(key[6])
	ptext[13] ^= byte(key[6] >> 8)
	ptext[14] ^= byte(key[7])
	ptext[15] ^= byte(key[7] >> 8)
}

/*
b. 	Byte Substitution Transformation: The Sub Byte transformation is a non-linear byte

	substitution that acts on every byte of the state in isolation to produce a new byte
	value, using S-box. This S-box is a simple table that contain a permutation of all
	possible 256 u8 values, and is used as a good shuffler for the bytes of the state.

	The proposed cipher is designed to have restrictions on the amount of ROM available,
	thus allowing the S-box to use only a small amount of memory with only 256 entries.
*/
func byteSubstitution(sbox *[256]byte, ptext []byte) {
	i := 0
	var row byte = 0
	var col byte = 0
	for i < 16 {
		row = ptext[i] & 15
		col = (ptext[i] >> 4) & 15
		// fmt.Printf("was: %d ", ptext[i])
		ptext[i] = sbox[(col<<4)+row]
		// fmt.Printf("now: %d\n", ptext[i])
		i += 1
	}
}

/*
c. 	Shift Rows Transformation: The action of shifting rows is particularly

	simple; just performing left circular shifts of rows 1, 2 and 3, by amounts of
	1, 2, and 3 bytes respectively. Row 0 is not changed. In the decryption process,
	the action of inverse shifting rows is particularly simple, just performing
	right circular shifts of rows 1, 2, and 3, by amounts of 1, 2, and 3 bytes,
	respectively.
*/
func shiftRows(ptext []byte) {
	row2 := binary.BigEndian.Uint32(ptext[:4])
	row3 := binary.BigEndian.Uint32(ptext[4:])
	row4 := binary.BigEndian.Uint32(ptext[8:])

	row2 = rotate(row2, 8)
	row3 = rotate(row3, 16)
	row4 = rotate(row4, 24)

	binary.LittleEndian.PutUint32(ptext[:4], row2)
	binary.LittleEndian.PutUint32(ptext[4:], row3)
	binary.LittleEndian.PutUint32(ptext[8:], row4)
}

/*
d. 	Dynamic Folding Transformation: In this transformation a complex rotation

	is applied to the state array by performing a dynamic permutation. In this
	stage, the elements of the state array are rearranged dynamically to new
	positions with more probabilities than they have in the normal arrangement.
	To perform the encryption/decryption process at the combiner, specific
	values are chosen from the keystream generated by the PRNG. The direction
	for the new permutation and the position to start are extracted from those
	values to implement the dynamic folding.

	The paper doesn't specify how to select the keystream values, only that specific
	values be chosen. So this scheme below is just one of many possibilities. All we
	do is sum adjacent bytes in the keystream and then mod by the remainder we need
	(first value mod 1 so our direction is between (0,0) and (1,1), and second value
	mod 3 to select a position in this 4x4 block)

	For a clearer depiction, see img/DYNAMIC_FOLD.png
*/
func dynamicFold(key *[8]uint32, ptext []byte, operation bool) {
	rotX := dynamicIdx(key, 0, 1)
	rotY := dynamicIdx(key, 2, 1)
	posX := dynamicIdx(key, 4, 3)
	posY := dynamicIdx(key, 6, 3)

	start := 0
	if posX >= 2 {
		start += 1
	}

	if posY <= 2 {
		start += 2
	}

	end := rotX + (rotY << 1)
	tmp := end - start
	if operation {
		if tmp == 1 || tmp == -1 {
			tmp = 3
		} else if tmp == 3 || tmp == -3 {
			tmp = 1
		}
	}

	if tmp < 0 {
		tmp *= -1
	}

	switch tmp & 3 {
	case 1:
		swap(ptext, 0, 8)
		swap(ptext, 1, 9)
		swap(ptext, 4, 12)
		swap(ptext, 5, 13)
		swap(ptext, 2, 10)
		swap(ptext, 3, 11)
		swap(ptext, 6, 14)
		swap(ptext, 7, 15)
		swap(ptext, 2, 8)
		swap(ptext, 3, 9)
		swap(ptext, 6, 12)
		swap(ptext, 7, 13)
	case 2:
		swap(ptext, 0, 10)
		swap(ptext, 1, 11)
		swap(ptext, 4, 14)
		swap(ptext, 5, 15)
		swap(ptext, 2, 8)
		swap(ptext, 3, 9)
		swap(ptext, 6, 12)
		swap(ptext, 7, 13)
	case 3:
		swap(ptext, 0, 2)
		swap(ptext, 1, 3)
		swap(ptext, 4, 6)
		swap(ptext, 5, 7)
		swap(ptext, 8, 10)
		swap(ptext, 9, 11)
		swap(ptext, 12, 14)
		swap(ptext, 13, 15)
		swap(ptext, 2, 8)
		swap(ptext, 3, 9)
		swap(ptext, 6, 12)
		swap(ptext, 7, 13)
	}
}

func combiner(key *[8]uint32, ptext []byte, sbox *[256]byte) {
	addRoundKs(key, ptext)
	byteSubstitution(sbox, ptext)
	shiftRows(ptext)
	dynamicFold(key, ptext, false)
}

func invCombiner(key *[8]uint32, ptext []byte, sbox *[256]byte) {
	dynamicFold(key, ptext, true)
	shiftRows(ptext)
	byteSubstitution(sbox, ptext)
	addRoundKs(key, ptext)
}
