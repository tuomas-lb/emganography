package framing

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	// Magic is the 4-byte magic identifier for the frame format
	Magic = "EMG0"
	// HeaderSize is the total size of the frame header in bytes
	HeaderSize = 16
	// CurrentVersion is the current frame format version
	CurrentVersion = 0x01
)

var (
	// ErrInvalidMagic indicates the frame magic bytes don't match
	ErrInvalidMagic = errors.New("invalid frame magic")
	// ErrInvalidLength indicates the payload length doesn't match the header
	ErrInvalidLength = errors.New("invalid payload length")
	// ErrCRCMismatch indicates the CRC32 checksum doesn't match
	ErrCRCMismatch = errors.New("CRC32 checksum mismatch")
	// ErrFrameTooShort indicates the frame is shorter than the header
	ErrFrameTooShort = errors.New("frame too short")
)

// Header represents the frame header structure
// Byte layout:
//   0-3:   Magic ("EMG0")
//   4:     Version (0x01)
//   5:     ECCScheme (1 byte)
//   6-7:   Reserved (0x00 0x00)
//   8-11:  PayloadLength (big-endian uint32)
//   12-15: PayloadCRC32 (big-endian CRC32-IEEE)
type Header struct {
	Magic         string
	Version       uint8
	ECCScheme     uint8
	Reserved      [2]byte
	PayloadLength uint32
	PayloadCRC32  uint32
}

// BuildFrame constructs a frame from a message and ECC scheme.
// The frame consists of: header (16 bytes) || message bytes
func BuildFrame(message []byte, eccScheme uint8) ([]byte, error) {
	// Calculate CRC32 of the message (payload only, no header)
	crc := crc32.ChecksumIEEE(message)

	// Build header
	header := make([]byte, HeaderSize)
	copy(header[0:4], []byte(Magic))
	header[4] = CurrentVersion
	header[5] = eccScheme
	// Reserved bytes [6-7] are already 0x00
	binary.BigEndian.PutUint32(header[8:12], uint32(len(message)))
	binary.BigEndian.PutUint32(header[12:16], crc)

	// Frame = header || message
	frame := make([]byte, HeaderSize+len(message))
	copy(frame[0:HeaderSize], header)
	copy(frame[HeaderSize:], message)

	return frame, nil
}

// ParseFrame parses a frame and validates its structure.
// Returns the header, payload bytes, and any error encountered.
func ParseFrame(frame []byte) (*Header, []byte, error) {
	if len(frame) < HeaderSize {
		return nil, nil, ErrFrameTooShort
	}

	// Extract magic
	magic := string(frame[0:4])
	if magic != Magic {
		return nil, nil, ErrInvalidMagic
	}

	// Extract header fields
	header := &Header{
		Magic:     magic,
		Version:   frame[4],
		ECCScheme: frame[5],
	}
	copy(header.Reserved[:], frame[6:8])
	header.PayloadLength = binary.BigEndian.Uint32(frame[8:12])
	header.PayloadCRC32 = binary.BigEndian.Uint32(frame[12:16])

	// Extract payload
	if len(frame) < HeaderSize+int(header.PayloadLength) {
		return nil, nil, ErrInvalidLength
	}
	payload := frame[HeaderSize : HeaderSize+int(header.PayloadLength)]

	// Validate CRC32
	calculatedCRC := crc32.ChecksumIEEE(payload)
	if calculatedCRC != header.PayloadCRC32 {
		return nil, nil, ErrCRCMismatch
	}

	return header, payload, nil
}



