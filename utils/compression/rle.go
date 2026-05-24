package compression

//zigzagOrder maps linear index in zigzag sequence to (row, col) in an 8x8 block
var zigzagOrder = [64][2]int{
	{0, 0}, {0, 1}, {1, 0}, {2, 0}, {1, 1}, {0, 2}, {0, 3}, {1, 2},
	{2, 1}, {3, 0}, {4, 0}, {3, 1}, {2, 2}, {1, 3}, {0, 4}, {0, 5},
	{1, 4}, {2, 3}, {3, 2}, {4, 1}, {5, 0}, {6, 0}, {5, 1}, {4, 2},
	{3, 3}, {2, 4}, {1, 5}, {0, 6}, {0, 7}, {1, 6}, {2, 5}, {3, 4},
	{4, 3}, {5, 2}, {6, 1}, {7, 0}, {7, 1}, {6, 2}, {5, 3}, {4, 4},
	{3, 5}, {2, 6}, {1, 7}, {2, 7}, {3, 6}, {4, 5}, {5, 4}, {6, 3},
	{7, 2}, {7, 3}, {6, 4}, {5, 5}, {4, 6}, {3, 7}, {4, 7}, {5, 6},
	{6, 5}, {7, 4}, {7, 5}, {6, 6}, {5, 7}, {6, 7}, {7, 6}, {7, 7},
}

func ZigzagScan(b Block) [64]int {
	var out [64]int
	for i, pos := range zigzagOrder {
		out[i] = int(b[pos[0]][pos[1]])
	}
	return out
}

//RLE encodes a zigzag-scanned block into a sequence of (skip, value) pairs, where skip counts the number of zeros before the next non-zero value.
func RLEEncode(zigzag [64]int) []RLEPair {
	pairs := []RLEPair{}
	zeros := 0

	pairs = append(pairs, RLEPair{Skip: 0, Value: zigzag[0]}) // DC coefficient is stored as-is

	for i := 1; i < 64; i++ {
		if zigzag[i] == 0 {
			zeros++
		} else {
			pairs = append(pairs, RLEPair{Skip: zeros, Value: zigzag[i]})
			zeros = 0
		}
	}

	pairs = append(pairs, RLEPair{Skip: 0, Value: 0}) // End of Block marker
	return pairs
}
