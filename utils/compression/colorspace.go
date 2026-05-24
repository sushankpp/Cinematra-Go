package compression

import "image"

//YCrCbPlane holds the three separated channel planes for a full frame
type YCbCrPlane struct {
	Y, Cb, Cr     [][]float64
	Width, Height int
}

//RCBToYCbCr converts an image's RGB pixels into three float64 planes.
// Formula is the standard BT.601 coefficients
func RGBToYCbCr(img image.Image) YCbCrPlane {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	Y := make([][]float64, h)
	Cb := make([][]float64, h)
	Cr := make([][]float64, h)

	for y := 0; y < h; y++ {
		Y[y] = make([]float64, w)
		Cb[y] = make([]float64, w)
		Cr[y] = make([]float64, w)

		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rf := float64(r >> 8)
			gf := float64(g >> 8)
			bf := float64(b >> 8)

			//BT.601 coefficients conversion
			Y[y][x] = 0.299*rf + 0.587*gf + 0.114*bf
			Cb[y][x] = -0.168736*rf - 0.331264*gf + 0.5*bf + 128
			Cr[y][x] = 0.5*rf - 0.418688*gf - 0.081312*bf + 128
		}
	}

	return YCbCrPlane{Y: Y, Cb: Cb, Cr: Cr, Width: w, Height: h}
}
