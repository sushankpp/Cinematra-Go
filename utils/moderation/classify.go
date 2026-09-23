package moderation

import "math"

type ClassificationResult struct {
	Label      string
	Confidence float64
	Scores     map[string]float64
}

// severeProfanities contains unambiguous vulgarities, slurs, and abusive profanities
// that should strictly never be allowed to be overridden by benign surrounding words (like "song", "video").
var severeProfanities = map[string]bool{
	"shit": true, "fuck": true, "fucking": true, "fucked": true, "bitch": true,
	"cunt": true, "asshole": true, "bastard": true, "retard": true, "retarded": true,
	"dick": true, "motherfucker": true, "dogshit": true, "dipshit": true,
	"pussy": true, "whore": true, "scumbag": true, "prick": true,
	// Nepali slurs & vulgarities
	"muji": true, "machikne": true, "machikney": true, "chikne": true, "lado": true,
	"kando": true, "radi": true, "randi": true, "jatha": true, "jatho": true,
	"chutiya": true, "madarchod": true, "bhenchod": true, "bhalu": true, "beshya": true,
}

func Classify(model *ContentFilterModel, text string) ClassificationResult {
	tokens := Tokenize(text)
	vocabSize := len(model.Vocabulary)

	// Filter tokens to only those that exist in the trained vocabulary.
	vocabSet := make(map[string]bool, vocabSize)
	for _, w := range model.Vocabulary {
		vocabSet[w] = true
	}

	var knownTokens []string
	hasProfanity := false
	for _, token := range tokens {
		if severeProfanities[token] {
			hasProfanity = true
		}
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

	// Direct profanities / slurs can never be masked by benign context
	if hasProfanity {
		logScores["Problematic"] += 25.0
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
