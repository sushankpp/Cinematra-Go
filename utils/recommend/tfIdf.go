package recommend

import (
	"math"
)

// this function takes all video documents and computes IDF across all video documents collection, then vectorizes every video. Returns a ready to use Corpus.
func BuildCorpus(docs []VideoDoc) *Corpus {
	totalVideos := len(docs)

	//tokenize every document once, keep the result
	tokensByVideo := make(map[string][]string)
	for _, doc := range docs {
		tokensByVideo[doc.VideoId] = Tokenize(doc.DocumentText)
	}

	//create document frequency
	df := make(map[string]int)
	for _, tokens := range tokensByVideo {
		seen := make(map[string]bool)
		for _, word := range tokens {
			if !seen[word] {
				seen[word] = true
				df[word]++
			}
		}
	}

	//computer idf for every word
	idf := make(map[string]float64)
	for word, count := range df {
		idf[word] = math.Log(float64(totalVideos) / float64(count))
	}

	//vectorize every video using its own term frequency * the shared IDF table
	videoVectors := make(map[string]VideoVector)
	for videoID, tokens := range tokensByVideo {
		vector := VectorizeTokens(tokens, idf)
		videoVectors[videoID] = VideoVector{
			VideoId: videoID,
			Vector:  vector,
		}
	}

	return &Corpus{
		VideoVectors: videoVectors,
		IDF:          idf,
	}
}

// VectorizeTokens converts a slice of tokens into a sparse TF-IDF vector using a given IDF table

func VectorizeTokens(tokens []string, idf map[string]float64) map[string]float64 {
	totalTokens := len(tokens)
	if totalTokens == 0 {
		return map[string]float64{}
	}

	//term frequency - raw count per word in THIS document
	tf := make(map[string]int)
	for _, word := range tokens {
		tf[word]++
	}

	vector := make(map[string]float64)
	for word, count := range tf {
		wordIDF, exists := idf[word]

		if !exists {
			continue
		}

		termFreq := float64(count) / float64(totalTokens)
		weight := termFreq * wordIDF

		if weight > 0 {
			vector[word] = weight
		}
	}

	return vector
}
