package ecc

import "errors"

// Scheme represents an error correction code scheme
type Scheme interface {
	// EncodeFrame encodes a frame (header + payload) into a bitstream
	EncodeFrame(frame []byte) ([]bool, error)
	// DecodeFrame decodes a bitstream into a frame (header + payload)
	// Returns the decoded frame bytes and any error encountered
	DecodeFrame(bits []bool) ([]byte, error)
}

// ECCScheme is an enum for different ECC schemes
type ECCScheme uint8

const (
	// ECCSchemeRepetition3 uses repetition-3 encoding (each bit repeated 3 times)
	ECCSchemeRepetition3 ECCScheme = 1
)

var (
	// ErrUnsupportedScheme indicates the ECC scheme is not supported
	ErrUnsupportedScheme = errors.New("unsupported ECC scheme")
	// ErrInsufficientBits indicates there are not enough bits to decode
	ErrInsufficientBits = errors.New("insufficient bits for decoding")
)

// GetScheme returns a Scheme implementation for the given ECCScheme
func GetScheme(scheme ECCScheme) (Scheme, error) {
	switch scheme {
	case ECCSchemeRepetition3:
		return &Repetition3{}, nil
	default:
		return nil, ErrUnsupportedScheme
	}
}



