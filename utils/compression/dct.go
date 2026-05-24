package compression

import "math"

var cosTable [8][8]float64

func init() {
	for n := 0; n < 8; n++ {
		for k := 0; k < 8; k++ {
			cosTable[n][k] = math.Cos((2*float64(n) + 1) * float64(k) * math.Pi / 16)
		}
	}
}

func alpha(k int) float64 {
	if k == 0 {
		return 1.0 / math.Sqrt2
	}

	return 1.0
}

// DCT applies the Discrete Cosine Transform to an 8x8 block of pixel values.
func DCT2D(b Block) Block {
	var rowPass Block

	for row := 0; row < 8; row++ {
		for k := 0; k < 8; k++ {
			sum := 0.0
			for n := 0; n < 8; n++ {
				sum += b[row][n] * cosTable[n][k]
			}
			rowPass[row][k] = 0.25 * alpha(k) * sum
		}
	}

	// apply 1D Dct to each column of the intermediate result
	var out Block
	for col := 0; col < 8; col++ {
		for k := 0; k < 8; k++ {
			sum := 0.0
			for n := 0; n < 8; n++ {
				sum += rowPass[n][col] * cosTable[n][k]
			}
			out[col][k] = 0.25 * alpha(k) * sum
		}
	}
	return out
}
