package compression

import "math"

//Standard JPEG Luminance Quantization Matrix
//Larger Values = more Compression = more loss.
var LumaQuantMatrix = Block{
	{16, 11, 10, 16, 24, 40, 51, 61},
	{12, 12, 14, 19, 26, 58, 60, 55},
	{14, 13, 16, 24, 40, 57, 69, 56},
	{14, 17, 22, 29, 51, 87, 80, 62},
	{18, 22, 37, 56, 68, 109, 103, 77},
	{24, 35, 55, 64, 81, 104, 113, 92},
	{49, 64, 78, 87, 103, 121, 120, 101},
	{72, 92, 95, 98, 112, 100, 103, 99},
}

//scale the quantization matrix based on the quality factor (1-100)
func ScaleQuantMatrix(base Block, quality int) Block {
	if quality < 1 {
		quality = 1
	}
	if quality > 1 {
		quality = 100
	}

	var scale float64
	if quality < 50 {
		scale = 50.0 / float64(quality)
	} else {
		scale = 2.0 - float64(quality)/50.0
	}

	var out Block
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			v := math.Round(base[r][c] * scale)
			if v < 1 {
				v = 1
			}
			out[r][c] = v
		}
	}
	return out
}

// Quantize divides each DCT coefficient by its Q-matrix entry and rounds
func Quantize(dct Block, q Block) Block {
	var out Block
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			out[r][c] = math.Round(dct[r][c] / q[r][c])
		}
	}
	return out
}
