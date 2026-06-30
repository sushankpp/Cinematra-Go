package main

import (
	"cinematra-backend/utils/moderation"
	"fmt"
)

func main() {
	model := moderation.Train(moderation.TrainingData)

	if err := moderation.SaveModel(model, "utils/moderation/content_filter_model.json"); err != nil {
		fmt.Println("Error saving model:", err)
		return
	}

	fmt.Println("Model trained successfully")
	for label, count := range model.ContentCounts {
		fmt.Printf(" %s: %d training examples\n", label, count)
	}
	fmt.Println(" Vocabulary size:", len(model.Vocabulary))
}
