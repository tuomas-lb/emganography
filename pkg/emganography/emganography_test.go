package emganography

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/tuomas-lb/emganography/internal/dct"
	"github.com/tuomas-lb/emganography/internal/imgutil"
	"github.com/tuomas-lb/emganography/internal/ycbcr"
)

// createTestImage creates a simple test image for testing
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a simple pattern
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8((x + y) * 255 / (width + height))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// saveTestImage saves a test image to a temporary file
func saveTestImage(t *testing.T, img *image.RGBA, filename string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, filename)

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return path
}

func TestEmbedExtractDCT_RoundTrip(t *testing.T) {
	// Create a test image (256x256 = 32x32 blocks = 1024 bits capacity)
	// With repetition-3: 1024/3 = 341 data bits = ~42 bytes
	// Header is 16 bytes, so we can embed ~26 bytes of data
	img := createTestImage(256, 256)
	testImagePath := saveTestImage(t, img, "test.png")

	// Test messages
	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "short ASCII",
			message: []byte("hello"),
		},
		{
			name:    "medium ASCII",
			message: []byte("Medium message test"),
		},
		{
			name:    "binary data",
			message: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultEmbedOptions()
			outputPath := filepath.Join(filepath.Dir(testImagePath), "output_"+tt.name+".png")

			// Embed
			err := EmbedMessageDCTFile(testImagePath, outputPath, tt.message, opts)
			if err != nil {
				t.Fatalf("EmbedMessageDCTFile failed: %v", err)
			}

			// Extract
			extracted, err := ExtractMessageDCTFile(outputPath)
			if err != nil {
				t.Fatalf("ExtractMessageDCTFile failed: %v", err)
			}

			// Compare
			if !bytes.Equal(tt.message, extracted) {
				t.Errorf("message mismatch: expected %v, got %v", tt.message, extracted)
			}
		})
	}
}

func TestEmbedExtractDCT_InMemory(t *testing.T) {
	// Create a test image (256x256 for sufficient capacity)
	img := createTestImage(256, 256)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
	inputData := buf.Bytes()

	message := []byte("test message")
	opts := DefaultEmbedOptions()

	// Embed
	outputData, err := EmbedMessageDCT(inputData, message, opts)
	if err != nil {
		t.Fatalf("EmbedMessageDCT failed: %v", err)
	}

	// Extract
	extracted, err := ExtractMessageDCT(outputData)
	if err != nil {
		t.Fatalf("ExtractMessageDCT failed: %v", err)
	}

	// Compare
	if !bytes.Equal(message, extracted) {
		t.Errorf("message mismatch: expected %v, got %v", message, extracted)
	}
}

func TestCapacityCheck(t *testing.T) {
	// Create a very small image (16x16 = 4 blocks = 4 bits capacity)
	// With repetition-3, that's only 1 bit of actual data capacity
	// But we need at least 16 bytes for the header, so this should fail
	img := createTestImage(16, 16)
	testImagePath := saveTestImage(t, img, "small.png")

	// Try to embed a message that's too long
	message := make([]byte, 100) // Way too long
	opts := DefaultEmbedOptions()
	outputPath := filepath.Join(filepath.Dir(testImagePath), "output.png")

	err := EmbedMessageDCTFile(testImagePath, outputPath, message, opts)
	if err != ErrMessageTooLong {
		t.Errorf("expected ErrMessageTooLong, got %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultDCTConfig()
	if config.ECC != ECCSchemeRepetition3 {
		t.Errorf("expected default ECC to be Repetition3")
	}
	if config.Delta == 0 {
		t.Errorf("expected default Delta to be non-zero")
	}
	if config.OutputFormat != "" {
		t.Errorf("expected default OutputFormat to be empty (preserves input format), got '%s'", config.OutputFormat)
	}
}

func TestImageLoadSaveRoundTrip(t *testing.T) {
	// Test loading and saving without any embedding or DCT processing
	// This should produce an identical image
	
	// Try to find test image
	var data []byte
	var err error
	paths := []string{
		"testdata/image.jpg",
		"testdata/image.png",
		"../../testdata/image.jpg",
		"../../testdata/image.png",
		"../testdata/image.jpg",
		"../testdata/image.png",
	}
	
	var foundPath string
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}
	if err != nil {
		t.Skipf("test image not found: %v", err)
	}
	
	t.Logf("Using test image: %s", foundPath)
	
	// Load original image
	img1, format1, err := imgutil.LoadImage(data)
	if err != nil {
		t.Fatalf("failed to load image: %v", err)
	}
	t.Logf("Loaded image format: %s, bounds: %v", format1, img1.Bounds())
	
	// Convert to YCbCr planes (no DCT processing)
	yPlane, cbPlane, crPlane := ycbcr.ImageToYCbCrPlanes(img1)
	
	// Convert back to image immediately (no modifications)
	img2 := ycbcr.YCbCrPlanesToImage(yPlane, cbPlane, crPlane)
	
	// Determine output format - use PNG for lossless comparison
	outputFormat := "png"
	if format1 == "jpeg" || format1 == "jpg" {
		// For JPEG, we'll test with PNG to ensure lossless round-trip
		// But also test with JPEG to see the loss
		outputFormat = "png"
	}
	
	// Save image
	outputData, err := imgutil.EncodeImage(img2, outputFormat, 100)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	
	// Save to temp file for inspection
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.png")
	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		t.Fatalf("failed to write output file: %v", err)
	}
	t.Logf("Saved output to: %s", outputPath)
	
	// Reload saved image
	img3, format3, err := imgutil.LoadImage(outputData)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	t.Logf("Reloaded image format: %s, bounds: %v", format3, img3.Bounds())
	
	// Compare pixel by pixel - compare RGB values directly
	// For PNG (lossless), the RGB values should be identical
	bounds2 := img2.Bounds()
	bounds3 := img3.Bounds()
	
	if bounds2.Dx() != bounds3.Dx() || bounds2.Dy() != bounds3.Dy() {
		t.Fatalf("image dimensions mismatch: saved %v, reloaded %v", bounds2, bounds3)
	}
	
	maxRDiff := 0
	maxGDiff := 0
	maxBDiff := 0
	differentPixels := 0
	totalPixels := bounds2.Dx() * bounds2.Dy()
	
	for y := 0; y < bounds2.Dy(); y++ {
		for x := 0; x < bounds2.Dx(); x++ {
			r1, g1, b1, a1 := img2.At(x, y).RGBA()
			r2, g2, b2, a2 := img3.At(x, y).RGBA()
			
			// Convert from 16-bit to 8-bit for comparison
			r1_8 := uint8(r1 >> 8)
			g1_8 := uint8(g1 >> 8)
			b1_8 := uint8(b1 >> 8)
			r2_8 := uint8(r2 >> 8)
			g2_8 := uint8(g2 >> 8)
			b2_8 := uint8(b2 >> 8)
			
			// Alpha should always be 255
			if a1 != a2 || a1 != 0xFFFF {
				t.Errorf("Alpha mismatch at (%d, %d): %d vs %d", x, y, a1, a2)
			}
			
			// Calculate differences
			diffR := int(r1_8) - int(r2_8)
			if diffR < 0 {
				diffR = -diffR
			}
			diffG := int(g1_8) - int(g2_8)
			if diffG < 0 {
				diffG = -diffG
			}
			diffB := int(b1_8) - int(b2_8)
			if diffB < 0 {
				diffB = -diffB
			}
			
			if diffR > maxRDiff {
				maxRDiff = diffR
			}
			if diffG > maxGDiff {
				maxGDiff = diffG
			}
			if diffB > maxBDiff {
				maxBDiff = diffB
			}
			
			if diffR > 0 || diffG > 0 || diffB > 0 {
				differentPixels++
			}
		}
	}
	
	t.Logf("RGB differences - R: %d, G: %d, B: %d", maxRDiff, maxGDiff, maxBDiff)
	t.Logf("Different pixels: %d / %d (%.2f%%)", 
		differentPixels, totalPixels, 100.0*float64(differentPixels)/float64(totalPixels))
	
	// For PNG (lossless), RGB values should be identical
	if outputFormat == "png" {
		if maxRDiff > 0 || maxGDiff > 0 || maxBDiff > 0 {
			t.Errorf("Images are not identical! Max differences - R: %d, G: %d, B: %d",
				maxRDiff, maxGDiff, maxBDiff)
		}
		if differentPixels > 0 {
			t.Errorf("Found %d different pixels in lossless format", differentPixels)
		}
	}
}

func TestRoundTripWithoutEmbedding(t *testing.T) {
	// Test the round-trip without any embedding to check for precision loss
	// Load image - try multiple paths
	var data []byte
	var err error
	paths := []string{"testdata/image.png", "../../testdata/image.png", "../testdata/image.png"}
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Skipf("test image not found: %v", err)
	}
	
	img1, _, err := imgutil.LoadImage(data)
	if err != nil {
		t.Fatalf("failed to load image: %v", err)
	}
	
	// Convert to YCbCr
	yPlane, cbPlane, crPlane := ycbcr.ImageToYCbCrPlanes(img1)
	
	// Process all 8x8 blocks: DCT -> IDCT without modification
	blocksAcross := yPlane.Width / 8
	blocksDown := yPlane.Height / 8
	
	var block [64]float64
	var dctBlock [64]float64
	
	for by := 0; by < blocksDown; by++ {
		for bx := 0; bx < blocksAcross; bx++ {
			// Extract 8x8 block and center
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					srcY := by*8 + y
					srcX := bx*8 + x
					block[y*8+x] = yPlane.Pix[srcY*yPlane.Stride+srcX] - 128.0
				}
			}
			
			// DCT
			dct.DCT8x8(&block, &dctBlock)
			
			// IDCT (no modification)
			dct.IDCT8x8(&dctBlock, &block)
			
			// Write back
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					srcY := by*8 + y
					srcX := bx*8 + x
					val := block[y*8+x] + 128.0
					if val < 0 { val = 0 }
					if val > 255 { val = 255 }
					yPlane.Pix[srcY*yPlane.Stride+srcX] = val
				}
			}
		}
	}
	
	// Convert back to image
	img2 := ycbcr.YCbCrPlanesToImage(yPlane, cbPlane, crPlane)
	
	// Save and reload
	outputData, err := imgutil.EncodeImage(img2, "png", 90)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	
	img3, _, err := imgutil.LoadImage(outputData)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	
	yPlane2, _, _ := ycbcr.ImageToYCbCrPlanes(img3)
	
	// Compare Y values
	maxDiff := 0.0
	differentPixels := 0
	totalPixels := yPlane.Width * yPlane.Height
	
	for i := 0; i < totalPixels; i++ {
		diff := yPlane.Pix[i] - yPlane2.Pix[i]
		if diff < 0 { diff = -diff }
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff > 0.5 {
			differentPixels++
		}
	}
	
	t.Logf("Max Y difference: %.6f", maxDiff)
	t.Logf("Pixels with difference > 0.5: %d / %d (%.2f%%)", 
		differentPixels, totalPixels, 100.0*float64(differentPixels)/float64(totalPixels))
	
	if maxDiff > 1.0 {
		t.Errorf("Significant precision loss detected! Max difference: %.6f", maxDiff)
	}
}

