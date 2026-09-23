package moderation

import (
	"testing"
)

func TestClassifySamples(t *testing.T) {
	if err := LoadGlobalModel("content_filter_model.json"); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	tests := []struct {
		text          string
		expectedLabel string
	}{
		// Safe
		{"What a fabulous song, truly melodious", "Safe"},
		{"what a masterpiece, pure cinema", "Safe"},
		{"fabulous singing and truly soothing melody", "Safe"},
		{"kasto mitho geet, sunera man shanta huncha", "Safe"},
		{"what is the name of this track?", "Safe"},
		{"कस्तो कर्णप्रिय र मिठो धुन, मनै शान्त भयो", "Safe"},

		// Potentially Problematic
		{"video dherai lamo bhayo ali boring lagyo", "PotentiallyProblematic"},
		{"the pacing was extremely slow and boring", "PotentiallyProblematic"},
		{"background music was too loud could barely hear your voice", "PotentiallyProblematic"},

		// Problematic
		{"This song is so so shit, don't every try to sing trash like this again", "Problematic"},
		{"what the fuck is this garbage", "Problematic"},
		{"shut the fuck up you piece of shit", "Problematic"},
		{"kasto muji video banako hawa", "Problematic"},
		{"radi ko ban kasto content ho yo", "Problematic"},
		{"this song is pure trash and absolute shit", "Problematic"},
		{"your singing is trash, stop embarrassing yourself", "Problematic"},
	}

	for _, tt := range tests {
		res := Classify(GlobalModel, tt.text)
		t.Logf("%-72s => [%s] (%.2f%%)", tt.text, res.Label, res.Confidence*100)
		if res.Label != tt.expectedLabel {
			t.Errorf("For text %q, expected label %q but got %q (%.2f%%)", tt.text, tt.expectedLabel, res.Label, res.Confidence*100)
		}
	}
}
