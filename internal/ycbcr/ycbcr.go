package ycbcr

import (
	"image"
	"image/color"
)

// Plane represents a 2D plane of float64 values with width, height, and stride
type Plane struct {
	Pix    []float64
	Width  int
	Height int
	Stride int
}

// ImageToYCbCrPlanes converts an image to Y, Cb, Cr planes
// Uses BT.601 coefficients for RGB to YCbCr conversion
// If the image is already YCbCr, preserves values directly
func ImageToYCbCrPlanes(img image.Image) (y, cb, cr *Plane) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	stride := width

	yPix := make([]float64, width*height)
	cbPix := make([]float64, width*height)
	crPix := make([]float64, width*height)

	// Convert from image to YCbCr planes
	// Handle YCbCr images specially to extract Y, Cb, Cr directly
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*stride + x
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			
			// Check if the color is already YCbCr
			if ycbcrColor, ok := c.(color.YCbCr); ok {
				// Extract Y, Cb, Cr directly from YCbCr color
				yPix[idx] = float64(ycbcrColor.Y)
				cbPix[idx] = float64(ycbcrColor.Cb)
				crPix[idx] = float64(ycbcrColor.Cr)
			} else {
				// Convert from RGB to YCbCr
				r, g, b, _ := c.RGBA()

				// Convert from 16-bit to 8-bit
				r8 := float64(r >> 8)
				g8 := float64(g >> 8)
				b8 := float64(b >> 8)

				// BT.601 coefficients
				// Y  = 0.299*R + 0.587*G + 0.114*B
				// Cb = -0.168736*R - 0.331264*G + 0.5*B + 128
				// Cr = 0.5*R - 0.418688*G - 0.081312*B + 128
				yPix[idx] = 0.299*r8 + 0.587*g8 + 0.114*b8
				cbPix[idx] = -0.168736*r8 - 0.331264*g8 + 0.5*b8 + 128.0
				crPix[idx] = 0.5*r8 - 0.418688*g8 - 0.081312*b8 + 128.0
			}
		}
	}

	return &Plane{Pix: yPix, Width: width, Height: height, Stride: stride},
		&Plane{Pix: cbPix, Width: width, Height: height, Stride: stride},
		&Plane{Pix: crPix, Width: width, Height: height, Stride: stride}
}

// YCbCrPlanesToImage converts Y, Cb, Cr planes back to an RGBA image
// Converts to RGBA explicitly to ensure consistent conversion when PNG encodes
func YCbCrPlanesToImage(y, cb, cr *Plane) *image.RGBA {
	width := y.Width
	height := y.Height
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for yIdx := 0; yIdx < height; yIdx++ {
		for xIdx := 0; xIdx < width; xIdx++ {
			idx := yIdx*y.Stride + xIdx

			Y := y.Pix[idx]
			Cb := cb.Pix[idx] - 128.0
			Cr := cr.Pix[idx] - 128.0

			// YCbCr to RGB conversion (BT.601)
			// R = Y + 1.402*Cr
			// G = Y - 0.344136*Cb - 0.714136*Cr
			// B = Y + 1.772*Cb
			r := Y + 1.402*Cr
			g := Y - 0.344136*Cb - 0.714136*Cr
			b := Y + 1.772*Cb

			// Clamp to [0, 255] and convert to uint8
			r8 := clamp(r)
			g8 := clamp(g)
			b8 := clamp(b)

			img.Set(xIdx, yIdx, color.RGBA{R: r8, G: g8, B: b8, A: 255})
		}
	}

	return img
}

// clamp clamps a float64 value to [0, 255] and returns as uint8
func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5) // Round
}

// clampToUint8 clamps a float64 value to [0, 255] and returns as uint8
func clampToUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5) // Round
}



