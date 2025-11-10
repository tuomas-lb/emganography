package framing

import (
	"testing"
)

func TestBuildFrame(t *testing.T) {
	message := []byte("hello")
	eccScheme := uint8(1)

	frame, err := BuildFrame(message, eccScheme)
	if err != nil {
		t.Fatalf("BuildFrame failed: %v", err)
	}

	if len(frame) != HeaderSize+len(message) {
		t.Errorf("expected frame length %d, got %d", HeaderSize+len(message), len(frame))
	}

	// Verify magic
	if string(frame[0:4]) != Magic {
		t.Errorf("expected magic %s, got %s", Magic, string(frame[0:4]))
	}
}

func TestParseFrame(t *testing.T) {
	message := []byte("hello")
	eccScheme := uint8(1)

	frame, err := BuildFrame(message, eccScheme)
	if err != nil {
		t.Fatalf("BuildFrame failed: %v", err)
	}

	header, payload, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame failed: %v", err)
	}

	if header.Magic != Magic {
		t.Errorf("expected magic %s, got %s", Magic, header.Magic)
	}

	if header.Version != CurrentVersion {
		t.Errorf("expected version %d, got %d", CurrentVersion, header.Version)
	}

	if header.ECCScheme != eccScheme {
		t.Errorf("expected ECC scheme %d, got %d", eccScheme, header.ECCScheme)
	}

	if header.PayloadLength != uint32(len(message)) {
		t.Errorf("expected payload length %d, got %d", len(message), header.PayloadLength)
	}

	if string(payload) != string(message) {
		t.Errorf("expected payload %s, got %s", message, payload)
	}
}

func TestParseFrame_InvalidMagic(t *testing.T) {
	frame := make([]byte, HeaderSize+10)
	copy(frame[0:4], []byte("XXXX"))

	_, _, err := ParseFrame(frame)
	if err != ErrInvalidMagic {
		t.Errorf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestParseFrame_CRCMismatch(t *testing.T) {
	message := []byte("hello")
	eccScheme := uint8(1)

	frame, err := BuildFrame(message, eccScheme)
	if err != nil {
		t.Fatalf("BuildFrame failed: %v", err)
	}

	// Corrupt the payload
	frame[HeaderSize] ^= 0xFF

	_, _, err = ParseFrame(frame)
	if err != ErrCRCMismatch {
		t.Errorf("expected ErrCRCMismatch, got %v", err)
	}
}



