package recommend

import "sync"

type VideoDoc struct {
	VideoId      string
	DocumentText string
}

type VideoVector struct {
	VideoId string
	Vector  map[string]float64
}

type Corpus struct {
	VideoVectors map[string]VideoVector
	IDF          map[string]float64
	mu           sync.RWMutex
}

type WatchedVideo struct {
	VideoId string
	Weight  float64
}
