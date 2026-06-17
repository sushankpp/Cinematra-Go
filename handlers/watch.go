package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

type WatchRequest struct {
	WatchDurationSeconds int  `json:"watch_duration_seconds"`
	Completed            bool `json:"completed"`
}

func WatchVideo(c fiber.Ctx) error {
	videoId := c.Params("id")
	userId := c.Locals("userID").(string)

	var req WatchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	type payload struct {
		UserId               string `json:"user_id"`
		VideoId              string `json:"video_id"`
		WatchDurationSeconds int    `json:"watch_duration_seconds"`
		Completed            bool   `json:"completed"`
		WatchedAt            string `json:"watched_at"`
	}

	body, err := json.Marshal(payload{
		UserId:               userId,
		VideoId:              videoId,
		WatchDurationSeconds: req.WatchDurationSeconds,
		Completed:            req.Completed,
		WatchedAt:            (time.Now().Format(time.RFC3339)),
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to build payload"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/watch_history?no_conflict=user_id,video_id"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to build payload"})
	}

	httpReq.Header.Set("Authorization", "Bearer "+serviceKey)
	httpReq.Header.Set("apikey", serviceKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save watch history"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "failed to save watch history", "details": string(respBody)})
	}

	return c.JSON(fiber.Map{"message": "watch history saved successfully", "success": true})
}

func GetWatchHistory(c fiber.Ctx) error {
	userId := c.Locals("userID").(string)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/watch_history?user_id=eq." + userId + "&select=*,videos(id,title,thumbnail_url,storage_url,language,content_rating)&order=watched_at.desc"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to build request"})
	}
	httpReq.Header.Set("Authorization", "Bearer "+serviceKey)
	httpReq.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch watch history"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch watch history", "details": string(respBody)})
	}

	respBody, _ := io.ReadAll(resp.Body)
	return c.Status(200).Type("json").Send(respBody)
}
