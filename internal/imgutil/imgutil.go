package imgutil

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

// LoadImageFromFile loads an image from a file path
// Returns the image, format string, and any error
func LoadImageFromFile(path string) (image.Image, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	return LoadImage(data)
}

// LoadImage loads an image from byte data
// Returns the image, format string, and any error
func LoadImage(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}
	return img, format, nil
}

// SaveImageToFile saves an image to a file
func SaveImageToFile(img image.Image, format, path string, quality int) error {
	data, err := EncodeImage(img, format, quality)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// EncodeImage encodes an image to the specified format
func EncodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	format = strings.ToLower(format)
	switch format {
	case "png", "image/png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %w", err)
		}
	case "jpg", "jpeg", "image/jpeg":
		opts := &jpeg.Options{Quality: quality}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return buf.Bytes(), nil
}

// CapacityBits calculates the number of bits that can be embedded in an image
// based on its dimensions (8x8 blocks)
func CapacityBits(width, height int) int {
	blocksAcross := width / 8
	blocksDown := height / 8
	return blocksAcross * blocksDown
}



