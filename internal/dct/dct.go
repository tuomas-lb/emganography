package dct

import "math"

// Precomputed cosine table for 8x8 DCT
// cosTable[i][j] = cos((2*i+1)*j*pi/16) for i,j in [0,7]
var cosTable [8][8]float64

func init() {
	// Precompute cosine values for 8x8 DCT
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			cosTable[i][j] = math.Cos(float64(2*i+1) * float64(j) * math.Pi / 16.0)
		}
	}
}

// DCT8x8 performs a 2D DCT on an 8x8 block
// Uses DCT-II formula: X[k] = C(k) * sum(n=0 to N-1) x[n] * cos((2n+1)kπ/(2N))
// where C(0) = sqrt(1/N), C(k) = sqrt(2/N) for k>0
// src and dst are 64-element arrays representing 8x8 blocks in row-major order
func DCT8x8(src *[64]float64, dst *[64]float64) {
	// First, apply 1D DCT to each row
	var temp [64]float64
	for row := 0; row < 8; row++ {
		for freq := 0; freq < 8; freq++ {
			sum := 0.0
			for col := 0; col < 8; col++ {
				sum += src[row*8+col] * cosTable[col][freq]
			}
			// Normalization: C(freq) = sqrt(1/8) for freq=0, sqrt(2/8) for freq>0
			c := math.Sqrt(2.0 / 8.0) // = 0.5
			if freq == 0 {
				c = math.Sqrt(1.0 / 8.0) // = 1/(2*sqrt(2)) ≈ 0.353553
			}
			temp[row*8+freq] = c * sum
		}
	}

	// Then, apply 1D DCT to each column (transform along rows for each column frequency)
	// temp[row*8+colFreq] contains row DCT results
	// We now transform along the row dimension for each column frequency
	for colFreq := 0; colFreq < 8; colFreq++ {
		for rowFreq := 0; rowFreq < 8; rowFreq++ {
			sum := 0.0
			for row := 0; row < 8; row++ {
				// temp[row*8+colFreq] is the value at row, column frequency colFreq
				sum += temp[row*8+colFreq] * cosTable[row][rowFreq]
			}
			// Normalization: C(rowFreq) = sqrt(1/8) for rowFreq=0, sqrt(2/8) for rowFreq>0
			c := math.Sqrt(2.0 / 8.0) // = 0.5
			if rowFreq == 0 {
				c = math.Sqrt(1.0 / 8.0) // = 1/(2*sqrt(2)) ≈ 0.353553
			}
			// Output: dst[rowFreq*8+colFreq] - row frequency first, then column frequency
			dst[rowFreq*8+colFreq] = c * sum
		}
	}
}

// IDCT8x8 performs a 2D inverse DCT on an 8x8 block
// Uses IDCT-II formula: x[n] = sum(k=0 to N-1) C(k) * X[k] * cos((2n+1)kπ/(2N))
// where C(0) = sqrt(1/N), C(k) = sqrt(2/N) for k>0
// src and dst are 64-element arrays representing 8x8 blocks in row-major order
// src is organized as [rowFreq*8+colFreq] from DCT output
func IDCT8x8(src *[64]float64, dst *[64]float64) {
	// First, apply 1D IDCT along columns (inverse transform along column frequency dimension)
	// This transforms each column from frequency domain back to spatial domain
	var temp [64]float64
	for colFreq := 0; colFreq < 8; colFreq++ {
		for row := 0; row < 8; row++ {
			sum := 0.0
			for rowFreq := 0; rowFreq < 8; rowFreq++ {
				// src[rowFreq*8+colFreq] - DCT coefficient at (rowFreq, colFreq)
				// Normalization: C(rowFreq) = sqrt(1/8) for rowFreq=0, sqrt(2/8) for rowFreq>0
				c := math.Sqrt(2.0 / 8.0) // = 0.5
				if rowFreq == 0 {
					c = math.Sqrt(1.0 / 8.0) // = 1/(2*sqrt(2)) ≈ 0.353553
				}
				sum += c * src[rowFreq*8+colFreq] * cosTable[row][rowFreq]
			}
			temp[row*8+colFreq] = sum
		}
	}

	// Then, apply 1D IDCT along rows (inverse transform along row frequency dimension)
	// This transforms each row from frequency domain back to spatial domain
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			sum := 0.0
			for colFreq := 0; colFreq < 8; colFreq++ {
				// temp[row*8+colFreq] contains intermediate results
				// Normalization: C(colFreq) = sqrt(1/8) for colFreq=0, sqrt(2/8) for colFreq>0
				c := math.Sqrt(2.0 / 8.0) // = 0.5
				if colFreq == 0 {
					c = math.Sqrt(1.0 / 8.0) // = 1/(2*sqrt(2)) ≈ 0.353553
				}
				sum += c * temp[row*8+colFreq] * cosTable[col][colFreq]
			}
			dst[row*8+col] = sum
		}
	}
}
