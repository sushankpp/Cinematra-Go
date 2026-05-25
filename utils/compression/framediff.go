package compression

import "image"

type FrameType int

const (
	IFrame FrameType = iota
	PFrame
)

type DiffBlock struct {
	Skip       bool
	Compressed CompressedBlock
}

type EncodedFrame struct {
	Type   FrameType
	Blocks [][]CompressedBlock
	Diff   [][]DiffBlock
}

type FrameDiffEncoder struct {
	prevY      [][]float64
	frameCount int
	iFrameRate int
	quality    int
}

func NewFrameDiffEncoder(iFrameRate, quality int) *FrameDiffEncoder {
	return &FrameDiffEncoder{
		iFrameRate: iFrameRate,
		quality:    quality,
	}
}

func (e *FrameDiffEncoder) Feed(img image.Image) EncodedFrame {
	planes := RGBToYCbCr(img)
	paddedY := PadToMultipleOf8(planes.Y)

	if e.frameCount == 0 || e.frameCount%e.iFrameRate == 0 {
		e.frameCount++
		e.prevY = paddedY
		return e.encodeIFrame(paddedY)
	}

	e.frameCount++
	diff := e.encodePFrame(paddedY)
	e.prevY = paddedY
	return diff
}

func (e *FrameDiffEncoder) encodeIFrame(Y [][]float64) EncodedFrame {
	rawBlocks := ExtractBlocks(Y)
	qMatrix := ScaleQuantMatrix(LumaQuantMatrix, e.quality)

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

	return EncodedFrame{Type: IFrame, Blocks: result}
}

func (e *FrameDiffEncoder) encodePFrame(currentY [][]float64) EncodedFrame {
	diffPlane := subtractPlanes(currentY, e.prevY)
	diffBlocks := ExtractBlocksRaw(diffPlane)
	qMatrix := ScaleQuantMatrix(LumaQuantMatrix, e.quality)

	rows := len(diffBlocks)
	result := make([][]DiffBlock, rows)

	for r, row := range diffBlocks {
		result[r] = make([]DiffBlock, len(row))
		for c, block := range row {
			if blockEnergy(block) < 64.0 {
				result[r][c] = DiffBlock{Skip: true}
				continue
			}

			freq := DCT2D(block)
			qBlock := Quantize(freq, qMatrix)
			seq := ZigzagScan(qBlock)
			pairs := RLEEncode(seq)

			result[r][c] = DiffBlock{
				Skip:       false,
				Compressed: CompressedBlock{Pairs: pairs},
			}
		}
	}

	return EncodedFrame{Type: PFrame, Diff: result}
}

func subtractPlanes(current, prev [][]float64) [][]float64 {
	h := len(current)
	w := len(current[0])

	diff := make([][]float64, h)
	for y := 0; y < h; y++ {
		diff[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			diff[y][x] = current[y][x] - prev[y][x]
		}
	}
	return diff
}

func blockEnergy(b Block) float64 {
	sum := 0.0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			v := b[r][c]
			if v < 0 {
				v = -v
			}
			sum += v
		}
	}
	return sum
}
