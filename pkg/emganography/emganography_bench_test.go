package emganography

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// createBenchmarkImage creates a test image for benchmarking
func createBenchmarkImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8((x + y) * 255 / (width + height))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func BenchmarkEmbedDCT_SmallImage(b *testing.B) {
	imageData := createBenchmarkImage(512, 512)
	message := []byte("This is a test message for benchmarking")
	opts := DefaultEmbedOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EmbedMessageDCT(imageData, message, opts)
		if err != nil {
			b.Fatalf("EmbedMessageDCT failed: %v", err)
		}
	}
}

func BenchmarkExtractDCT_SmallImage(b *testing.B) {
	imageData := createBenchmarkImage(512, 512)
	message := []byte("This is a test message for benchmarking")
	opts := DefaultEmbedOptions()

	// Embed first
	embeddedData, err := EmbedMessageDCT(imageData, message, opts)
	if err != nil {
		b.Fatalf("EmbedMessageDCT failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExtractMessageDCT(embeddedData)
		if err != nil {
			b.Fatalf("ExtractMessageDCT failed: %v", err)
		}
	}
}

func BenchmarkEmbedDCT_MediumImage(b *testing.B) {
	imageData := createBenchmarkImage(1024, 1024)
	message := []byte("This is a longer test message for benchmarking medium-sized images")
	opts := DefaultEmbedOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EmbedMessageDCT(imageData, message, opts)
		if err != nil {
			b.Fatalf("EmbedMessageDCT failed: %v", err)
		}
	}
}

func BenchmarkExtractDCT_MediumImage(b *testing.B) {
	imageData := createBenchmarkImage(1024, 1024)
	message := []byte("This is a longer test message for benchmarking medium-sized images")
	opts := DefaultEmbedOptions()

	// Embed first
	embeddedData, err := EmbedMessageDCT(imageData, message, opts)
	if err != nil {
		b.Fatalf("EmbedMessageDCT failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExtractMessageDCT(embeddedData)
		if err != nil {
			b.Fatalf("ExtractMessageDCT failed: %v", err)
		}
	}
}



