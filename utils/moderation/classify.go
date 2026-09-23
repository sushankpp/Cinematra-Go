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

	// Filter tokens to only those that exist in the trained vocabulary.
	// Out-Of-Vocabulary (OOV) tokens (e.g. unknown actor names, proper nouns)
	// contain zero discriminative evidence. Applying Laplace smoothing to dozens
	// of unseen words artificially skews towards whichever class has fewer total words.
	vocabSet := make(map[string]bool, vocabSize)
	for _, w := range model.Vocabulary {
		vocabSet[w] = true
	}

	var knownTokens []string
	for _, token := range tokens {
		if vocabSet[token] {
			knownTokens = append(knownTokens, token)
		}
	}

	// If no known words exist in the text, it has zero toxic or problematic signal -> Safe
	if len(knownTokens) == 0 {
		return ClassificationResult{
			Label:      "Safe",
			Confidence: 1.0,
			Scores:     map[string]float64{"Safe": 0.0},
		}
	}

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

		for _, word := range knownTokens {
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
