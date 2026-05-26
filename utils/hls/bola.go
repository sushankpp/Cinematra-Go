package hls

import (
	"fmt"
	"math"
)

type BolaState struct {
	Qualities    []Quality
	V            float64
	Gamma        float64
	BufferTarget float64
}

type QualityDecision struct {
	Quality      Quality
	QualityIndex int
	Score        float64
	BufferLevel  float64
}

func NewBOLAState(qualities []Quality, gamma, bufferTarget float64) *BolaState {
	v := (bufferTarget - float64(len(qualities))) / (utilityAt(qualities, len(qualities)-1) + 1.0/float64(len(qualities)))

	return &BolaState{
		Qualities:    qualities,
		V:            v,
		Gamma:        1.0,
		BufferTarget: bufferTarget,
	}
}

func (b *BolaState) NextQuality(bufferLevel float64) QualityDecision {
	bestScore := math.Inf(-1)
	bestIndex := 0

	for i := range b.Qualities {
		score := b.score(i, bufferLevel)
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}

	return QualityDecision{
		Quality:      b.Qualities[bestIndex],
		QualityIndex: bestIndex,
		Score:        bestScore,
		BufferLevel:  bufferLevel,
	}
}

func (b *BolaState) score(qualityIndex int, bufferLevel float64) float64 {
	u := utilityAt(b.Qualities, qualityIndex)
	size := bitrateKbps(b.Qualities[qualityIndex])
	numerator := b.V*u + b.V*b.Gamma/b.BufferTarget - bufferLevel
	return numerator / size
}

func utilityAt(qualities []Quality, index int) float64 {
	if index == 0 {
		return 0
	}

	lowestBitrate := bitrateKbps(qualities[0])
	thisBitrate := bitrateKbps(qualities[index])
	return math.Log(thisBitrate / lowestBitrate)
}

func bitrateKbps(q Quality) float64 {
	var n float64
	fmt.Sscanf(q.Bitrate, "%fk", &n)
	return n
}
