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
