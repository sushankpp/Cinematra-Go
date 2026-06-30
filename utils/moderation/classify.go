package moderation

import "math"

type ClassificationResult struct {
	Label      string
	Confidence float64
	Scores     map[string]float64
}

func Classify(model *ContentFilterModel, text string) ClassificationResult {
	tokens := Tokenize(text)
	vocabSize := len(model.Vocabulary)

	totalDocs := 0
	for _, count := range model.ContentCounts {
		totalDocs += count
	}

	logScores := make(map[string]float64)

	for label, classDocCount := range model.ContentCounts {
		prior := float64(classDocCount) / float64(totalDocs)
		logScore := math.Log(prior)

		totalWordsInClass := 0
		for _, c := range model.WordFrequencies[label] {
			totalWordsInClass += c
		}

		for _, word := range tokens {
			wordCount := model.WordFrequencies[label][word]
			wordProb := float64(wordCount+1) / float64(totalWordsInClass+vocabSize)
			logScore += math.Log(wordProb)
		}

		logScores[label] = logScore
	}

	bestLabel := ""
	bestScore := math.Inf(-1)
	for label, score := range logScores {
		if score > bestScore {
			bestScore = score
			bestLabel = label
		}
	}

	maxLog := bestScore
	sumExp := 0.0
	for _, score := range logScores {
		sumExp += math.Exp(score - maxLog)
	}
	confidence := 1.0 / sumExp

	return ClassificationResult{
		Label:      bestLabel,
		Confidence: confidence,
		Scores:     logScores,
	}
}
