package moderation

import (
	"encoding/json"
	"os"
)

type ContentFilterModel struct {
	WordFrequencies map[string]map[string]int `json:"WordFrequencies"`
	ContentCounts   map[string]int            `json:"ContentCounts"`
	Vocabulary      []string                  `json:"Vocabulary"`
}

var GlobalModel *ContentFilterModel

func SaveModel(model *ContentFilterModel, path string) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadModel(path string) (*ContentFilterModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model ContentFilterModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func LoadGlobalModel(path string) error {
	model, err := LoadModel(path)
	if err != nil {
		return err
	}
	GlobalModel = model
	return nil
}
