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

// HybridScoredVideo holds a video ID and its component similarity, trending, and combined scores.
type HybridScoredVideo struct {
	VideoId       string
	Similarity    float64
	TrendingScore float64
	CombinedScore float64
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

// RecommendHybrid scores candidates using alpha * similarity + beta * normalizedTrending,
// and returns them sorted by combined score descending.
// Unwatched videos are prioritized first. If unwatched candidates are fewer than limit,
// the remaining slots are backfilled with the user's highest-affinity / trending watched videos.
func RecommendHybrid(
	userProfile map[string]float64,
	corpus *Corpus,
	trendingScores map[string]float64,
	maxTrending float64,
	watchedIDs map[string]bool,
	alpha, beta float64,
	limit int,
) []HybridScoredVideo {
	unwatchedCandidates := []HybridScoredVideo{}
	watchedCandidates := []HybridScoredVideo{}

	for videoId, videoVector := range corpus.VideoVectors {
		similarity := 0.0
		if len(userProfile) > 0 {
			similarity = CosineSimilarity(userProfile, videoVector.Vector)
		}

		trending := 0.0
		if maxTrending > 0 {
			if s, exists := trendingScores[videoId]; exists {
				trending = s / maxTrending
			}
		}

		combined := alpha*similarity + beta*trending
		if combined > 0 {
			candidate := HybridScoredVideo{
				VideoId:       videoId,
				Similarity:    similarity,
				TrendingScore: trending,
				CombinedScore: combined,
			}
			if watchedIDs != nil && watchedIDs[videoId] {
				watchedCandidates = append(watchedCandidates, candidate)
			} else {
				unwatchedCandidates = append(unwatchedCandidates, candidate)
			}
		}
	}

	sort.Slice(unwatchedCandidates, func(i, j int) bool {
		return unwatchedCandidates[i].CombinedScore > unwatchedCandidates[j].CombinedScore
	})

	results := unwatchedCandidates
	if len(results) > limit {
		results = results[:limit]
	}

	// Backfill with top-scoring watched videos if unwatched catalog is exhausted
	if len(results) < limit && len(watchedCandidates) > 0 {
		sort.Slice(watchedCandidates, func(i, j int) bool {
			return watchedCandidates[i].CombinedScore > watchedCandidates[j].CombinedScore
		})
		needed := limit - len(results)
		if needed > len(watchedCandidates) {
			needed = len(watchedCandidates)
		}
		results = append(results, watchedCandidates[:needed]...)
	}

	return results
}
