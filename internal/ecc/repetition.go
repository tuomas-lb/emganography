package ecc

import (
	"errors"

	"github.com/tuomas-lb/emganography/internal/bitstream"
)

var (
	// ErrCorruptedTriple indicates a triple of bits couldn't be decoded (all 3 differ)
	ErrCorruptedTriple = errors.New("corrupted triple: all bits differ")
)

// Repetition3 implements repetition-3 error correction coding
// Each data bit is encoded as 3 identical bits (b, b, b)
// Decoding uses majority vote on each triple
type Repetition3 struct{}

// EncodeFrame encodes a frame into a bitstream using repetition-3
func (r *Repetition3) EncodeFrame(frame []byte) ([]bool, error) {
	// Convert frame bytes to bits
	dataBits := bitstream.BytesToBits(frame)

	// Encode each bit as 3 identical bits
	encodedBits := make([]bool, 0, len(dataBits)*3)
	for _, bit := range dataBits {
		encodedBits = append(encodedBits, bit, bit, bit)
	}

	return encodedBits, nil
}

// DecodeFrame decodes a bitstream using repetition-3 majority voting
func (r *Repetition3) DecodeFrame(bits []bool) ([]byte, error) {
	if len(bits) == 0 {
		return nil, ErrInsufficientBits
	}

	// Group bits into triples and decode using majority vote
	tripleCount := len(bits) / 3
	if tripleCount == 0 {
		return nil, ErrInsufficientBits
	}

	decodedBits := make([]bool, tripleCount)
	for i := 0; i < tripleCount; i++ {
		offset := i * 3
		b1, b2, b3 := bits[offset], bits[offset+1], bits[offset+2]

		// Majority vote
		ones := 0
		if b1 {
			ones++
		}
		if b2 {
			ones++
		}
		if b3 {
			ones++
		}

		if ones >= 2 {
			decodedBits[i] = true
		} else {
			decodedBits[i] = false
		}
	}

	// Convert decoded bits back to bytes
	return bitstream.BitsToBytes(decodedBits), nil
}



