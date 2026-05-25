package compression

// ExtractBlocks returns a 2D slice of Block — each element is already [8][8]float64
func ExtractBlocks(plane [][]float64) [][]Block {
	h := len(plane)
	if h == 0 {
		return nil
	}
	w := len(plane[0])

	blockRows := h / 8
	blockCols := w / 8

	blocks := make([][]Block, blockRows) // ← [][]Block, not [][][8]float64
	for br := 0; br < blockRows; br++ {
		blocks[br] = make([]Block, blockCols) // ← []Block
		for bc := 0; bc < blockCols; bc++ {
			var b Block // ← Block = [8][8]float64
			for row := 0; row < 8; row++ {
				for col := 0; col < 8; col++ {
					b[row][col] = plane[br*8+row][bc*8+col] - 128
				}
			}
			blocks[br][bc] = b
		}
	}
	return blocks
}

// PadToMultipleOf8 extends a plane to the nearest 8-pixel boundary
// by repeating the last row/column
func PadToMultipleOf8(plane [][]float64) [][]float64 {
	h := len(plane)
	if h == 0 {
		return plane
	}
	w := len(plane[0])

	newH := (h + 7) / 8 * 8
	newW := (w + 7) / 8 * 8

	// if already exact multiples, return as-is
	if newH == h && newW == w {
		return plane
	}

	padded := make([][]float64, newH)
	for r := 0; r < newH; r++ {
		padded[r] = make([]float64, newW)
		srcRow := r
		if srcRow >= h {
			srcRow = h - 1 // repeat last row
		}
		for c := 0; c < newW; c++ {
			srcCol := c
			if srcCol >= w {
				srcCol = w - 1 // repeat last column
			}
			padded[r][c] = plane[srcRow][srcCol]
		}
	}
	return padded
}

// ExtractBlocksRaw splits a plane into 8×8 blocks WITHOUT level shifting.
// Used for diff planes which are already centered around zero.
func ExtractBlocksRaw(plane [][]float64) [][]Block {
	h := len(plane)
	if h == 0 {
		return nil
	}
	w := len(plane[0])

	blockRows := h / 8
	blockCols := w / 8

	blocks := make([][]Block, blockRows)
	for br := 0; br < blockRows; br++ {
		blocks[br] = make([]Block, blockCols)
		for bc := 0; bc < blockCols; bc++ {
			var b Block
			for row := 0; row < 8; row++ {
				for col := 0; col < 8; col++ {
					b[row][col] = plane[br*8+row][bc*8+col] // no -128
				}
			}
			blocks[br][bc] = b
		}
	}
	return blocks
}
