package moderation

func Train(data []LabeledComment) *ContentFilterModel {
	model := &ContentFilterModel{
		WordFrequencies: make(map[string]map[string]int),
		ContentCounts:   make(map[string]int),
	}

	vocabSet := make(map[string]bool)

	for _, item := range data {
		model.ContentCounts[item.Label]++

		if model.WordFrequencies[item.Label] == nil {
			model.WordFrequencies[item.Label] = make(map[string]int)
		}

		for _, word := range Tokenize(item.Text) {
			model.WordFrequencies[item.Label][word]++
			vocabSet[word] = true
		}
	}

	for word := range vocabSet {
		model.Vocabulary = append(model.Vocabulary, word)
	}
	return model
}
