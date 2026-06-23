package recommend

import (
	"math"
	"sort"
)

// ScoredVideo holds a video ID and its computed similarity score.
type ScoredVideo struct {
	VideoId    string
	Similarity float64
}

// CosineSimilarity computes the cosine similarity between two word-score vectors.
func CosineSimilarity(a, b map[string]float64) float64 {
	var dot, magA, magB float64

	// Calculate dot product and magnitude of vector a
	for word, scoreA := range a {
		magA += scoreA * scoreA
		if scoreB, exists := b[word]; exists {
			dot += scoreA * scoreB
		}
	}

	// Calculate magnitude of vector b
	for _, scoreB := range b {
		magB += scoreB * scoreB
	}

	// Prevent division by zero
	if magA == 0.0 || magB == 0.0 {
		return 0.0
	}

	return dot / math.Sqrt(magA*magB)
}

// Recommend returns scored videos sorted by cosine similarity (descending).
func Recommend(userProfile map[string]float64, corpus *Corpus, watchedIDs map[string]bool, limit int) []ScoredVideo {
	candidates := []ScoredVideo{}

	for videoId, videoVector := range corpus.VideoVectors {
		if watchedIDs[videoId] {
			continue
		}

		similarity := CosineSimilarity(userProfile, videoVector.Vector)
		if similarity > 0 {
			candidates = append(candidates, ScoredVideo{
				VideoId:    videoId,
				Similarity: similarity,
			})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Similarity > candidates[j].Similarity
	})

	if len(candidates) < limit {
		limit = len(candidates)
	}

	return candidates[:limit]
}
