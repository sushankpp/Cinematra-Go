package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

func ReactToVideo(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	videoID := c.Params("id")

	var body struct {
		ReactionType string `json:"reaction_type"` // "like" or "dislike"
	}
	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.ReactionType != "like" && body.ReactionType != "dislike" {
		return c.Status(400).JSON(fiber.Map{"error": "reaction_type must be like or dislike"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	client := &http.Client{Timeout: 10 * time.Second}

	// check existing reaction
	checkURL := fmt.Sprintf("%s/rest/v1/video_reactions?video_id=eq.%s&user_id=eq.%s&select=reaction_type", supabaseURL, videoID, userID)
	checkReq, _ := http.NewRequest("GET", checkURL, nil)
	checkReq.Header.Set("apikey", serviceKey)
	checkReq.Header.Set("Authorization", "Bearer "+serviceKey)
	checkReq.Header.Set("Accept", "application/vnd.pgrst.object+json")

	checkResp, err := client.Do(checkReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer checkResp.Body.Close()

	if checkResp.StatusCode == 200 {
		// existing reaction found
		var existing struct {
			ReactionType string `json:"reaction_type"`
		}
		checkBody, _ := io.ReadAll(checkResp.Body)
		json.Unmarshal(checkBody, &existing)

		if existing.ReactionType == body.ReactionType {
			// same reaction tapped again -> remove
			delURL := fmt.Sprintf("%s/rest/v1/video_reactions?video_id=eq.%s&user_id=eq.%s", supabaseURL, videoID, userID)
			delReq, _ := http.NewRequest("DELETE", delURL, nil)
			delReq.Header.Set("apikey", serviceKey)
			delReq.Header.Set("Authorization", "Bearer "+serviceKey)
			resp, err := client.Do(delReq)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			defer resp.Body.Close()
			return c.Status(200).JSON(fiber.Map{"message": "reaction removed", "reaction": nil, "user_reaction": nil})
		}

		// different reaction -> update
		updURL := fmt.Sprintf("%s/rest/v1/video_reactions?video_id=eq.%s&user_id=eq.%s", supabaseURL, videoID, userID)
		payload, _ := json.Marshal(map[string]string{"reaction_type": body.ReactionType})
		updReq, _ := http.NewRequest("PATCH", updURL, bytes.NewBuffer(payload))
		updReq.Header.Set("apikey", serviceKey)
		updReq.Header.Set("Authorization", "Bearer "+serviceKey)
		updReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(updReq)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer resp.Body.Close()
		return c.Status(200).JSON(fiber.Map{"message": "reaction updated", "reaction": body.ReactionType, "user_reaction": body.ReactionType})
	}

	// no existing reaction -> insert
	insURL := supabaseURL + "/rest/v1/video_reactions"
	payload, _ := json.Marshal(map[string]string{
		"video_id":      videoID,
		"user_id":       userID,
		"reaction_type": body.ReactionType,
	})
	insReq, _ := http.NewRequest("POST", insURL, bytes.NewBuffer(payload))
	insReq.Header.Set("apikey", serviceKey)
	insReq.Header.Set("Authorization", "Bearer "+serviceKey)
	insReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(insReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "failed to save reaction: " + string(respBody)})
	}
	return c.Status(201).JSON(fiber.Map{"message": "reaction added", "reaction": body.ReactionType, "user_reaction": body.ReactionType})
}
