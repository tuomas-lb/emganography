package bitstream

import (
	"reflect"
	"testing"
)

func TestBytesToBits(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []bool
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: nil,
		},
		{
			name:     "single byte 0x00",
			input:    []byte{0x00},
			expected: []bool{false, false, false, false, false, false, false, false},
		},
		{
			name:     "single byte 0xFF",
			input:    []byte{0xFF},
			expected: []bool{true, true, true, true, true, true, true, true},
		},
		{
			name:     "single byte 0x80",
			input:    []byte{0x80},
			expected: []bool{true, false, false, false, false, false, false, false},
		},
		{
			name:     "two bytes",
			input:    []byte{0x80, 0x01},
			expected: []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToBits(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBitsToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []bool
		expected []byte
	}{
		{
			name:     "empty",
			input:    []bool{},
			expected: nil,
		},
		{
			name:     "8 bits all false",
			input:    []bool{false, false, false, false, false, false, false, false},
			expected: []byte{0x00},
		},
		{
			name:     "8 bits all true",
			input:    []bool{true, true, true, true, true, true, true, true},
			expected: []byte{0xFF},
		},
		{
			name:     "8 bits MSB set",
			input:    []bool{true, false, false, false, false, false, false, false},
			expected: []byte{0x80},
		},
		{
			name:     "16 bits",
			input:    []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true},
			expected: []byte{0x80, 0x01},
		},
		{
			name:     "7 bits (padded)",
			input:    []bool{true, false, false, false, false, false, false},
			expected: []byte{0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BitsToBytes(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	bits := BytesToBits(original)
	result := BitsToBytes(bits)

	if !reflect.DeepEqual(original, result) {
		t.Errorf("round trip failed: expected %v, got %v", original, result)
	}
}



