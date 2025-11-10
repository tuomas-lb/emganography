package emganography

import (
	"errors"
	"fmt"
	"image"
	"os"

	"github.com/tuomas-lb/emganography/internal/dct"
	"github.com/tuomas-lb/emganography/internal/ecc"
	"github.com/tuomas-lb/emganography/internal/framing"
	"github.com/tuomas-lb/emganography/internal/imgutil"
	"github.com/tuomas-lb/emganography/internal/ycbcr"
)

var (
	// ErrMessageTooLong indicates the message exceeds the image capacity
	ErrMessageTooLong = errors.New("message too long for image capacity")
	// ErrFrameCorrupt indicates the extracted frame is corrupted
	ErrFrameCorrupt = errors.New("extracted frame is corrupted")
	// ErrCRCMismatch indicates CRC validation failed
	ErrCRCMismatch = errors.New("CRC32 checksum mismatch")
)

// CapacityInfo holds information about image embedding capacity
type CapacityInfo struct {
	// Image dimensions
	Width  int
	Height int
	// Raw capacity in 8x8 blocks
	BlocksAcross int
	BlocksDown   int
	// Capacity in bits (number of 8x8 blocks)
	CapacityBits int
	// Maximum embeddable payload bytes (after accounting for header and ECC)
	MaxPayloadBytes int
	// Maximum embeddable UTF-8 string length (approximate)
	MaxUTF8Chars int
}

// ECCScheme represents an error correction code scheme
type ECCScheme = ecc.ECCScheme

const (
	// ECCSchemeRepetition3 uses repetition-3 encoding
	ECCSchemeRepetition3 = ecc.ECCSchemeRepetition3
)

// DCTConfig holds configuration for DCT-based embedding
type DCTConfig struct {
	// ECC is the error correction scheme to use
	ECC ECCScheme
	// Delta is the coefficient adjustment magnitude
	Delta float64
	// MinGap is the minimum required difference between coeffs to encode a bit
	MinGap float64
	// UseAllBlocks if true, use all blocks; else allow skipping low-energy blocks
	UseAllBlocks bool
	// OutputFormat is the output image format: "png" or "jpg"
	OutputFormat string
}

// DefaultDCTConfig returns a default DCT configuration
func DefaultDCTConfig() DCTConfig {
	return DCTConfig{
		ECC:          ECCSchemeRepetition3,
		Delta:        10.0, // Reduced from 200.0 for less visible artifacts
		MinGap:       5.0,  // Reduced from 100.0 for less visible artifacts
		UseAllBlocks: true,
		OutputFormat: "", // Empty means preserve input format
	}
}

// EmbedOptions holds options for embedding
type EmbedOptions struct {
	// Config is the DCT configuration
	Config DCTConfig
	// JPEGQuality is the JPEG quality (1-100) if output format is JPEG, default 90
	JPEGQuality int
}

// DefaultEmbedOptions returns default embedding options
func DefaultEmbedOptions() *EmbedOptions {
	return &EmbedOptions{
		Config:      DefaultDCTConfig(),
		JPEGQuality: 90,
	}
}

// EmbedMessageDCTFile embeds a message into an image file using DCT
func EmbedMessageDCTFile(inputPath, outputPath string, message []byte, opts *EmbedOptions) error {
	if opts == nil {
		opts = DefaultEmbedOptions()
	}

	// Load image data
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Embed in memory
	outputData, err := EmbedMessageDCT(inputData, message, opts)
	if err != nil {
		return err
	}

	// Save embedded image
	return os.WriteFile(outputPath, outputData, 0644)
}

// EmbedMessageDCT embeds a message into an image using DCT
// input is the encoded image bytes (PNG/JPEG), or nil to load from file
// Returns encoded image bytes with embedded message
func EmbedMessageDCT(input []byte, message []byte, opts *EmbedOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultEmbedOptions()
	}

	// Load image
	var img image.Image
	var format string
	var err error
	if input == nil {
		return nil, fmt.Errorf("input data required")
	}
	img, format, err = imgutil.LoadImage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Convert to YCbCr planes
	yPlane, cbPlane, crPlane := ycbcr.ImageToYCbCrPlanes(img)

	// Build frame (header + message)
	frame, err := framing.BuildFrame(message, uint8(opts.Config.ECC))
	if err != nil {
		return nil, fmt.Errorf("failed to build frame: %w", err)
	}

	// Get ECC scheme
	eccScheme, err := ecc.GetScheme(opts.Config.ECC)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECC scheme: %w", err)
	}

	// ECC encode frame
	encodedBits, err := eccScheme.EncodeFrame(frame)
	if err != nil {
		return nil, fmt.Errorf("failed to ECC encode: %w", err)
	}

	// Check capacity
	capacityBits := imgutil.CapacityBits(yPlane.Width, yPlane.Height)
	if len(encodedBits) > capacityBits {
		return nil, ErrMessageTooLong
	}

	// Embed bits into DCT coefficients
	err = embedBitsIntoDCT(yPlane, encodedBits, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to embed bits: %w", err)
	}

	// Convert back to image
	outputImg := ycbcr.YCbCrPlanesToImage(yPlane, cbPlane, crPlane)

	// Determine output format
	outputFormat := opts.Config.OutputFormat
	if outputFormat == "" {
		outputFormat = format
	}
	if outputFormat == "" {
		outputFormat = "png"
	}

	// Encode image
	return imgutil.EncodeImage(outputImg, outputFormat, opts.JPEGQuality)
}

// ExtractMessageDCTFile extracts a message from an image file using DCT
func ExtractMessageDCTFile(inputPath string) ([]byte, error) {
	// Load image data
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ExtractMessageDCT(data)
}

// ExtractMessageDCT extracts a message from an image using DCT
func ExtractMessageDCT(input []byte) ([]byte, error) {
	// Load image
	img, _, err := imgutil.LoadImage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Convert to YCbCr planes
	yPlane, _, _ := ycbcr.ImageToYCbCrPlanes(img)

	// Determine ECC scheme (assume Repetition3 for now)
	eccScheme, err := ecc.GetScheme(ECCSchemeRepetition3)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECC scheme: %w", err)
	}

	capacityBits := imgutil.CapacityBits(yPlane.Width, yPlane.Height)

	// First pass: Extract enough bits to decode the frame header
	// Header is 16 bytes = 128 bits, with repetition-3 = 384 encoded bits
	minBitsForHeader := framing.HeaderSize * 8 * 3 // 384 bits
	extractBits := minBitsForHeader
	if capacityBits < extractBits {
		extractBits = capacityBits
	}

	extractedBits := extractBitsFromDCT(yPlane, extractBits)
	frameBytes, err := eccScheme.DecodeFrame(extractedBits)
	if err != nil {
		return nil, fmt.Errorf("failed to ECC decode header: %w", err)
	}

	// Parse header to get payload length
	if len(frameBytes) < framing.HeaderSize {
		// Not enough bits, try extracting more
		extractedBits = extractBitsFromDCT(yPlane, capacityBits)
		frameBytes, err = eccScheme.DecodeFrame(extractedBits)
		if err != nil {
			return nil, fmt.Errorf("failed to ECC decode: %w", err)
		}
	}

	// Try to parse just the header fields manually to get payload length
	if len(frameBytes) < framing.HeaderSize {
		return nil, fmt.Errorf("insufficient data for frame header")
	}

	// Validate magic bytes first
	magic := string(frameBytes[0:4])
	if magic != framing.Magic {
		// Header is corrupted, try extracting all bits
		extractedBits = extractBitsFromDCT(yPlane, capacityBits)
		frameBytes, err = eccScheme.DecodeFrame(extractedBits)
		if err != nil {
			return nil, fmt.Errorf("failed to ECC decode: %w", err)
		}
		// Try parsing again
		header, payload, err := framing.ParseFrame(frameBytes)
		if err != nil {
			if errors.Is(err, framing.ErrCRCMismatch) {
				return nil, ErrCRCMismatch
			}
			return nil, fmt.Errorf("%w: %v", ErrFrameCorrupt, err)
		}
		if header.ECCScheme != uint8(ECCSchemeRepetition3) {
			eccScheme, err = ecc.GetScheme(ECCScheme(header.ECCScheme))
			if err != nil {
				return nil, fmt.Errorf("unsupported ECC scheme: %d", header.ECCScheme)
			}
			frameBytes, err = eccScheme.DecodeFrame(extractedBits)
			if err != nil {
				return nil, fmt.Errorf("failed to ECC decode with correct scheme: %w", err)
			}
			_, payload, err = framing.ParseFrame(frameBytes)
			if err != nil {
				if errors.Is(err, framing.ErrCRCMismatch) {
					return nil, ErrCRCMismatch
				}
				return nil, fmt.Errorf("%w: %v", ErrFrameCorrupt, err)
			}
		}
		return payload, nil
	}

	// Read payload length from header (bytes 8-11, big-endian uint32)
	payloadLength := uint32(frameBytes[8])<<24 | uint32(frameBytes[9])<<16 | uint32(frameBytes[10])<<8 | uint32(frameBytes[11])
	// Sanity check payload length
	if payloadLength > 1000000 { // Unreasonably large
		return nil, fmt.Errorf("invalid payload length in header: %d", payloadLength)
	}
	totalFrameBytes := framing.HeaderSize + int(payloadLength)
	totalFrameBits := totalFrameBytes * 8 * 3 // With repetition-3

	// Second pass: Extract exactly the number of bits needed for the full frame
	if totalFrameBits > capacityBits {
		return nil, fmt.Errorf("frame requires %d bits but capacity is only %d", totalFrameBits, capacityBits)
	}

	extractedBits = extractBitsFromDCT(yPlane, totalFrameBits)
	frameBytes, err = eccScheme.DecodeFrame(extractedBits)
	if err != nil {
		return nil, fmt.Errorf("failed to ECC decode full frame: %w", err)
	}

	// Parse frame
	header, payload, err := framing.ParseFrame(frameBytes)
	if err != nil {
		if errors.Is(err, framing.ErrCRCMismatch) {
			return nil, ErrCRCMismatch
		}
		return nil, fmt.Errorf("%w: %v", ErrFrameCorrupt, err)
	}

	// Verify ECC scheme matches
	if header.ECCScheme != uint8(ECCSchemeRepetition3) {
		// Try with the correct scheme
		eccScheme, err = ecc.GetScheme(ECCScheme(header.ECCScheme))
		if err != nil {
			return nil, fmt.Errorf("unsupported ECC scheme in frame: %d", header.ECCScheme)
		}
		// Re-decode with correct scheme
		frameBytes, err = eccScheme.DecodeFrame(extractedBits)
		if err != nil {
			return nil, fmt.Errorf("failed to ECC decode with correct scheme: %w", err)
		}
		_, payload, err = framing.ParseFrame(frameBytes)
		if err != nil {
			if errors.Is(err, framing.ErrCRCMismatch) {
				return nil, ErrCRCMismatch
			}
			return nil, fmt.Errorf("%w: %v", ErrFrameCorrupt, err)
		}
	}

	return payload, nil
}

// GetCapacityInfoFromData calculates capacity from image data in memory
func GetCapacityInfoFromData(data []byte, eccScheme ECCScheme) (*CapacityInfo, error) {
	img, _, err := imgutil.LoadImage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Convert to YCbCr to get dimensions
	yPlane, _, _ := ycbcr.ImageToYCbCrPlanes(img)

	// Calculate capacity
	width := yPlane.Width
	height := yPlane.Height
	blocksAcross := width / 8
	blocksDown := height / 8
	capacityBits := blocksAcross * blocksDown

	// Get ECC scheme to determine expansion factor
	ecc, err := ecc.GetScheme(eccScheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECC scheme: %w", err)
	}

	// Calculate maximum payload
	// Frame = header (16 bytes) + payload
	// Encoded bits = frameBytes * 8 * eccExpansion
	// We need: capacityBits >= (16 + payloadBytes) * 8 * eccExpansion
	// Solving: payloadBytes <= (capacityBits / (8 * eccExpansion)) - 16

	// For repetition-3, expansion is 3
	// Test with a dummy frame to get the expansion factor
	testFrame := make([]byte, framing.HeaderSize+1) // 16 + 1 = 17 bytes
	encodedBits, err := ecc.EncodeFrame(testFrame)
	if err != nil {
		return nil, fmt.Errorf("failed to encode test frame: %w", err)
	}

	// Calculate expansion factor
	expansionFactor := len(encodedBits) / (len(testFrame) * 8)

	// Calculate max payload bytes
	maxFrameBytes := capacityBits / (8 * expansionFactor)
	maxPayloadBytes := maxFrameBytes - framing.HeaderSize
	if maxPayloadBytes < 0 {
		maxPayloadBytes = 0
	}

	// Estimate UTF-8 character capacity (most UTF-8 chars are 1 byte, but some are 2-4)
	// Use a conservative estimate: assume average 1.5 bytes per UTF-8 char
	maxUTF8Chars := int(float64(maxPayloadBytes) / 1.5)

	return &CapacityInfo{
		Width:           width,
		Height:          height,
		BlocksAcross:    blocksAcross,
		BlocksDown:      blocksDown,
		CapacityBits:    capacityBits,
		MaxPayloadBytes: maxPayloadBytes,
		MaxUTF8Chars:    maxUTF8Chars,
	}, nil
}

// GetCapacityInfo calculates capacity from an image file
func GetCapacityInfo(inputPath string, eccScheme ECCScheme) (*CapacityInfo, error) {
	// Load image
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return GetCapacityInfoFromData(data, eccScheme)
}

// embedBitsIntoDCT embeds bits into DCT coefficients of Y plane
func embedBitsIntoDCT(yPlane *ycbcr.Plane, bits []bool, config DCTConfig) error {
	blocksAcross := yPlane.Width / 8
	blocksDown := yPlane.Height / 8
	bitIdx := 0

	var block [64]float64
	var dctBlock [64]float64

	for by := 0; by < blocksDown; by++ {
		for bx := 0; bx < blocksAcross; bx++ {
			// Extract 8x8 block and center values (subtract 128) for DCT
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					srcY := by*8 + y
					srcX := bx*8 + x
					block[y*8+x] = yPlane.Pix[srcY*yPlane.Stride+srcX] - 128.0
				}
			}

			// Apply DCT
			dct.DCT8x8(&block, &dctBlock)

			// Embed bit if available
			if bitIdx < len(bits) {
				bit := bits[bitIdx]
				coeff22 := dctBlock[2*8+2] // (2,2)
				coeff23 := dctBlock[2*8+3] // (2,3)

				// Adjust coefficients symmetrically to encode bit
				// Only modify (2,2) and (2,3), no other coefficients
				// Always enforce the relationship to ensure reliable extraction
				midpoint := (coeff22 + coeff23) / 2.0
				requiredGap := config.MinGap + config.Delta

				if bit {
					// Encode 1: ensure (2,2) > (2,3) by at least MinGap
					dctBlock[2*8+2] = midpoint + requiredGap/2.0
					dctBlock[2*8+3] = midpoint - requiredGap/2.0
				} else {
					// Encode 0: ensure (2,2) < (2,3) by at least MinGap
					dctBlock[2*8+2] = midpoint - requiredGap/2.0
					dctBlock[2*8+3] = midpoint + requiredGap/2.0
				}
				bitIdx++
			}

			// Apply inverse DCT
			dct.IDCT8x8(&dctBlock, &block)

			// Write back to Y plane with clamping (add 128 back after IDCT)
			// Keep as float64 to preserve precision through the round-trip
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					srcY := by*8 + y
					srcX := bx*8 + x
					val := block[y*8+x] + 128.0
					if val < 0 {
						val = 0
					}
					if val > 255 {
						val = 255
					}
					// Keep as float64, don't round yet - rounding happens in YCbCr->RGB conversion
					yPlane.Pix[srcY*yPlane.Stride+srcX] = val
				}
			}
		}
	}

	return nil
}

// extractBitsFromDCT extracts bits from DCT coefficients of Y plane
func extractBitsFromDCT(yPlane *ycbcr.Plane, maxBits int) []bool {
	blocksAcross := yPlane.Width / 8
	blocksDown := yPlane.Height / 8
	bits := make([]bool, 0, maxBits)

	var block [64]float64
	var dctBlock [64]float64

	for by := 0; by < blocksDown && len(bits) < maxBits; by++ {
		for bx := 0; bx < blocksAcross && len(bits) < maxBits; bx++ {
			// Extract 8x8 block and center values (subtract 128) for DCT
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					srcY := by*8 + y
					srcX := bx*8 + x
					block[y*8+x] = yPlane.Pix[srcY*yPlane.Stride+srcX] - 128.0
				}
			}

			// Apply DCT
			dct.DCT8x8(&block, &dctBlock)

			// Extract bit by comparing coefficients
			coeff22 := dctBlock[2*8+2] // (2,2)
			coeff23 := dctBlock[2*8+3] // (2,3)

			bit := coeff22 > coeff23
			bits = append(bits, bit)
		}
	}

	return bits
}
