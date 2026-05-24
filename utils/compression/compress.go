package compression

import "image"

func CompressFrame(img image.Image, quality int) [][]CompressedBlock {
	planes := RGBToYCbCr(img)

	rawBlocks := ExtractBlocks(planes.Y)

	qMatrix := ScaleQuantMatrix(LumaQuantMatrix, quality)

	rows := len(rawBlocks)
	result := make([][]CompressedBlock, rows)

	for r, row := range rawBlocks {
		result[r] = make([]CompressedBlock, len(row))

		for c, block := range row {
			freq := DCT2D(block)
			qBlock := Quantize(freq, qMatrix)
			seq := ZigzagScan(qBlock)
			pairs := RLEEncode(seq)
			result[r][c] = CompressedBlock{Pairs: pairs}
		}
	}

	return result
}
