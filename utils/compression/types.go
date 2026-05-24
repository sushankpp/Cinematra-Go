package compression

// Block represents an 8x8 block of pixel values, used for image compression and DCT operations.
type Block [8][8]float64

//RLEPair is one run-legnth encoded symbol:
// Skip = number of zeros before this value
//value = the actual non-zero coefficient
type RLEPair struct {
	Skip  int
	Value int
}

// Compressed Block is the final output for 8x8 Block
type CompressedBlock struct {
	Pairs []RLEPair
}
