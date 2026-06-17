package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v3"
)

type PreferenceRequest struct {
	PreferredGenres    []int    `json:"preferred_genres"`
	PreferredTags      []int    `json:"preferred_tags"`
	PreferredLanguages []string `json:"preferred_languages"`
	PreferredRatings   []string `json:"preferred_ratings"`
}

func SaveUserPreference(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req PreferenceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	type payload struct {
		UserID             string   `json:"user_id"`
		PreferredGenres    []int    `json:"preferred_genres"`
		PreferredTags      []int    `json:"preferred_tags"`
		PreferredLanguages []string `json:"preferred_languages"`
		PreferredRatings   []string `json:"preferred_ratings"`
	}

	body, err := json.Marshal(payload{
		UserID:             userID,
		PreferredGenres:    req.PreferredGenres,
		PreferredTags:      req.PreferredTags,
		PreferredLanguages: req.PreferredLanguages,
		PreferredRatings:   req.PreferredRatings,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to build payload"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/user_preferences"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	httpReq.Header.Set("Authorization", "Bearer "+serviceKey)
	httpReq.Header.Set("apikey", serviceKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "failed to save preferences: " + string(respBody)})
	}

	return c.JSON(fiber.Map{"message": "preferences saved successfully", "success": true})
}

func GetUserPreference(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := fmt.Sprintf("%s/rest/v1/user_preferences?user_id=eq.%s", supabaseURL, userID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	httpReq.Header.Set("Authorization", "Bearer "+serviceKey)
	httpReq.Header.Set("apikey", serviceKey)
	httpReq.Header.Set("Accept", "application/vnd.pgrst.object+json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return c.Status(404).JSON(fiber.Map{"error": "preferences not found"})
	}

	respBody, _ := io.ReadAll(resp.Body)
	return c.Status(200).Type("json").Send(respBody)
}
