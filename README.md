# Emganography-Go

A high-performance Go library for DCT-based steganography with robust error correction.

## Overview

This library implements steganography using Discrete Cosine Transform (DCT) on 8×8 blocks of the Y channel (luminance) in YCbCr color space. It features:

- **DCT-based embedding**: Embeds data in DCT coefficients (2,2) and (2,3) of 8×8 blocks
- **Error correction**: Hybrid ECC scheme with repetition-3 encoding
- **Framing**: Structured frame format with magic bytes, version, CRC32 checksum
- **Performance**: Optimized for low allocations and high throughput

## Installation

```bash
go get github.com/tuomas-lb/emganography
```

## Quick Start

### Using the Library

```go
package main

import (
    "fmt"
    "os"
    "github.com/tuomass/emganography-go/pkg/emganography"
)

func main() {
    // Embed a message (file-based)
    message := []byte("Hello, steganography!")
    opts := emganography.DefaultEmbedOptions()
    
    err := emganography.EmbedMessageDCTFile("input.jpg", "output.jpg", message, opts)
    if err != nil {
        panic(err)
    }
    
    // Extract the message (file-based)
    extracted, err := emganography.ExtractMessageDCTFile("output.jpg")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Extracted: %s\n", string(extracted))
    
    // In-memory embedding
    inputData, _ := os.ReadFile("input.jpg")
    outputData, err := emganography.EmbedMessageDCT(inputData, message, opts)
    if err != nil {
        panic(err)
    }
    os.WriteFile("output.jpg", outputData, 0644)
    
    // Check capacity
    info, err := emganography.GetCapacityInfo("input.jpg", emganography.ECCSchemeRepetition3)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Max payload: %d bytes\n", info.MaxPayloadBytes)
    fmt.Printf("Max UTF-8 chars: ~%d\n", info.MaxUTF8Chars)
}
```

### Using the CLI Tool

```bash
# Build the CLI
go build -o emgtool ./cmd/emgtool

# Embed a message
./emgtool embed-dct -in input.jpg -out output.jpg -msg "Hello, world!"

# Embed from a file
./emgtool embed-dct -in input.png -out output.png -msg-file message.txt

# Extract a message (prints to stdout)
./emgtool extract-dct -in output.jpg

# Extract to a file
./emgtool extract-dct -in output.jpg -out extracted.txt

# Check image capacity (maximum embeddable UTF-8 characters)
./emgtool capacity -in image.jpg

# Get detailed capacity information
./emgtool capacity -in image.jpg -format both

# Use stdin/stdout for piping
cat image.jpg | ./emgtool capacity -in -
cat image.jpg | ./emgtool embed-dct -in - -out - -msg "Hello" > output.jpg
```

## CLI Commands

### `embed-dct`

Embeds a message into an image using DCT-based steganography.

**Flags:**
- `-in <path>`: Input image path (use `-` for stdin)
- `-out <path>`: Output image path (use `-` for stdout)
- `-msg <text>`: Message to embed (or use `-msg-file`)
- `-msg-file <path>`: File containing message to embed
- `-ecc <scheme>`: ECC scheme (default: `repetition3`)
- `-format <format>`: Output format (`png` or `jpg`, default: preserve input format)
- `-quality <1-100>`: JPEG quality (default: 90, only for JPEG output)

**Examples:**
```bash
# Embed a text message
./emgtool embed-dct -in photo.jpg -out secret.jpg -msg "Secret message"

# Embed from file, preserve format
./emgtool embed-dct -in photo.jpg -out secret.jpg -msg-file data.bin

# Convert to PNG
./emgtool embed-dct -in photo.jpg -out secret.png -msg "Hello" -format png

# Pipe input/output
cat photo.jpg | ./emgtool embed-dct -in - -out - -msg "Hello" > output.jpg
```

### `extract-dct`

Extracts a message from an image.

**Flags:**
- `-in <path>`: Input image path (use `-` for stdin)
- `-out <path>`: Output file path (optional, prints to stdout if not set)

**Examples:**
```bash
# Extract to stdout
./emgtool extract-dct -in secret.jpg

# Extract to file
./emgtool extract-dct -in secret.jpg -out message.txt

# Pipe input
cat secret.jpg | ./emgtool extract-dct -in -
```

### `capacity`

Inspects an image and shows the maximum embeddable data capacity.

**Flags:**
- `-in <path>`: Input image path (use `-` for stdin)
- `-ecc <scheme>`: ECC scheme (default: `repetition3`)
- `-format <format>`: Output format: `utf8` (default), `bytes`, or `both`

**Examples:**
```bash
# Get UTF-8 character capacity (default)
./emgtool capacity -in image.jpg
# Output: 444

# Get byte capacity
./emgtool capacity -in image.jpg -format bytes
# Output: 666

# Get detailed information
./emgtool capacity -in image.jpg -format both
# Output:
# Image: 1024x1024 pixels
# Blocks: 128x128 (16384 total blocks)
# Capacity: 16384 bits
# Maximum embeddable data: 666 bytes
# Maximum UTF-8 string length: ~444 characters

# Pipe input
cat image.jpg | ./emgtool capacity -in -
```

## Architecture

The library is organized into several internal packages:

- **`internal/framing`**: Frame construction and parsing with CRC32 validation
- **`internal/ecc`**: Error correction code implementations (repetition-3)
- **`internal/bitstream`**: Bit-level conversions between bytes and bits
- **`internal/dct`**: 2D DCT/IDCT implementation for 8×8 blocks
- **`internal/ycbcr`**: RGB to YCbCr conversion utilities
- **`internal/imgutil`**: Image loading, saving, and capacity calculation
- **`pkg/emganography`**: Public API for embedding and extraction

## Frame Format

The library uses a structured frame format:

```
Header (16 bytes):
  - Magic: 4 bytes ("EMG0")
  - Version: 1 byte (0x01)
  - ECCScheme: 1 byte
  - Reserved: 2 bytes
  - PayloadLength: 4 bytes (big-endian)
  - PayloadCRC32: 4 bytes (big-endian)

Frame = Header || Payload
```

The frame is then ECC-encoded (repetition-3) before embedding into the image.

## Features

- **Format Support**: Works with both PNG and JPEG images
- **Format Preservation**: By default, preserves the input image format
- **Stdin/Stdout Support**: All commands support piping with `-in -` and `-out -`
- **Capacity Inspection**: Check maximum embeddable data before embedding
- **Error Correction**: Repetition-3 ECC for robust message recovery
- **Frame Validation**: CRC32 checksum ensures message integrity
- **Low Artifacts**: Optimized DCT coefficient modification for minimal visual impact

## Capacity

Image capacity is calculated as:
- `blocksAcross = width / 8`
- `blocksDown = height / 8`
- `capacityBits = blocksAcross * blocksDown`

With repetition-3 ECC, the actual data capacity is `capacityBits / 3` bits, minus the 16-byte header overhead.

The `capacity` command provides:
- **Raw capacity**: Total number of bits available (one per 8×8 block)
- **Maximum payload bytes**: Maximum embeddable data after accounting for ECC expansion and header
- **UTF-8 character estimate**: Conservative estimate of maximum UTF-8 string length (~1.5 bytes per character average)

## Testing

Run all tests:
```bash
go test ./...
```

Run benchmarks:
```bash
go test -bench=. ./pkg/emganography -benchmem
```

## License

This project is provided as-is for educational and research purposes.



