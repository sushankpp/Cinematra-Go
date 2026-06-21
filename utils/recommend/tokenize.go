package recommend

import (
	"strings"
)

// tokenize.go converts the raw text of video documents into a vector representation. Converts raw text into slice of clean, lowercase words, removing punctuation and stop words.
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	rawWords := strings.Split(text, " ")

	var tokens []string
	for _, w := range rawWords {
		w = strings.Trim(w, ".,!?;:\"()[]{}") // Remove punctuation
		if (len(w) < 2) || stopWords[w] {
			continue
		}

		tokens = append(tokens, w)
	}

	return tokens
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"is": true, "are": true, "was": true, "were": true, "in": true,
	"on": true, "at": true, "to": true, "of": true, "for": true,
	"with": true, "this": true, "that": true, "it": true, "as": true,
}
