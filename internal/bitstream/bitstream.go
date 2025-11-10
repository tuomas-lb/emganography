package bitstream

// BytesToBits converts a byte slice to a boolean slice representing bits.
// Each byte is converted to 8 bits, MSB first.
func BytesToBits(data []byte) []bool {
	if len(data) == 0 {
		return nil
	}
	bits := make([]bool, len(data)*8)
	for i, b := range data {
		offset := i * 8
		for j := 0; j < 8; j++ {
			bits[offset+j] = (b>>(7-j))&1 == 1
		}
	}
	return bits
}

// BitsToBytes converts a boolean slice to a byte slice.
// Bits are packed MSB first, with any trailing bits padded with zeros.
func BitsToBytes(bits []bool) []byte {
	if len(bits) == 0 {
		return nil
	}
	byteCount := (len(bits) + 7) / 8
	bytes := make([]byte, byteCount)
	for i, bit := range bits {
		if bit {
			byteIdx := i / 8
			bitIdx := 7 - (i % 8)
			bytes[byteIdx] |= 1 << bitIdx
		}
	}
	return bytes
}



