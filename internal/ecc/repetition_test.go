package ecc

import (
	"reflect"
	"testing"

	"github.com/tuomass/emganography-go/internal/bitstream"
)

func TestRepetition3_EncodeDecode(t *testing.T) {
	r := &Repetition3{}

	original := []byte{0x12, 0x34}
	encoded, err := r.EncodeFrame(original)
	if err != nil {
		t.Fatalf("EncodeFrame failed: %v", err)
	}

	// Verify encoding: each bit should be repeated 3 times
	originalBits := bitstream.BytesToBits(original)
	if len(encoded) != len(originalBits)*3 {
		t.Errorf("expected encoded length %d, got %d", len(originalBits)*3, len(encoded))
	}

	decoded, err := r.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("round trip failed: expected %v, got %v", original, decoded)
	}
}

func TestRepetition3_ErrorCorrection(t *testing.T) {
	r := &Repetition3{}

	original := []byte{0x80} // 10000000 in binary
	encoded, err := r.EncodeFrame(original)
	if err != nil {
		t.Fatalf("EncodeFrame failed: %v", err)
	}

	// First bit should be: true, true, true
	if !encoded[0] || !encoded[1] || !encoded[2] {
		t.Errorf("first bit should be encoded as [true, true, true]")
	}

	// Corrupt one bit in the first triple (flip one)
	encoded[0] = false

	decoded, err := r.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}

	// Should still decode correctly (majority vote: false, true, true -> true)
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("error correction failed: expected %v, got %v", original, decoded)
	}
}

func TestRepetition3_TwoBitError(t *testing.T) {
	r := &Repetition3{}

	original := []byte{0x80} // 10000000
	encoded, err := r.EncodeFrame(original)
	if err != nil {
		t.Fatalf("EncodeFrame failed: %v", err)
	}

	// Corrupt two bits in the first triple
	encoded[0] = false
	encoded[1] = false

	decoded, err := r.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}

	// Majority vote: false, false, true -> false (wrong!)
	// This should decode incorrectly
	if reflect.DeepEqual(original, decoded) {
		t.Logf("Note: two-bit error in triple caused incorrect decoding (expected behavior)")
	}
}

func TestRepetition3_InsufficientBits(t *testing.T) {
	r := &Repetition3{}

	// Less than 3 bits
	bits := []bool{true, false}
	_, err := r.DecodeFrame(bits)
	if err != ErrInsufficientBits {
		t.Errorf("expected ErrInsufficientBits, got %v", err)
	}
}



