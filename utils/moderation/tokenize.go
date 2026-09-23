package moderation

import (
	"strings"
	"unicode"
)

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"is": true, "are": true, "am": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true, "by": true, "from": true,
	"this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
	"and": true, "or": true, "as": true, "if": true,
	"what": true, "which": true, "who": true, "whom": true, "where": true, "when": true, "how": true,
	"so": true, "very": true, "too": true, "really": true, "every": true, "all": true, "some": true, "any": true, "just": true,
	"like": true, "again": true, "don": true, "t": true, "s": true, "m": true, "d": true, "ll": true, "ve": true, "re": true,
	"do": true, "does": true, "did": true, "can": true, "could": true, "will": true, "would": true, "have": true, "has": true, "had": true,
	"my": true, "your": true, "his": true, "her": true, "their": true, "our": true, "me": true, "you": true, "we": true, "they": true, "them": true,
	// common romanized nepali grammatical particles
	"ko": true, "le": true, "ma": true, "ra": true, "lai": true, "ni": true, "hai": true, "ta": true, "yo": true, "tyo": true, "pani": true, "ani": true,
}

func Tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				word := current.String()
				if !stopWords[word] {
					tokens = append(tokens, word)
				}
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		word := current.String()
		if !stopWords[word] {
			tokens = append(tokens, word)
		}
	}
	return tokens
}
