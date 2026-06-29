package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Subscribe(c fiber.Ctx) error {
	subscriberID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		ChannelID string `json:"channel_id"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.ChannelID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "channel_id is required"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload, err := json.Marshal(map[string]string{
		"subscriber_id": subscriberID,
		"channel_id":    body.ChannelID,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	url := supabaseURL + "/rest/v1/subscriptions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return c.Status(500).JSON(fiber.Map{"error": "failed to subscribe: " + string(respBody)})
	}

	return c.Status(201).JSON(fiber.Map{"message": "subscribed successfully"})
}

func UnSubscribe(c fiber.Ctx) error {
	subscriberID, ok := c.Locals("userID").(string)

	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	channelID := c.Params("channelId")
	if channelID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "channel_id is required"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := fmt.Sprintf(supabaseURL+"/rest/v1/subscriptions?subscriber_id=eq.%s&channel_id=eq.%s", subscriberID, channelID)
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 204 {
		return c.Status(500).JSON(fiber.Map{"error": "failed to unsubscribe: " + string(respBody), "status": resp.StatusCode})
	}

	return c.Status(200).JSON(fiber.Map{"message": "unsubscribed successfully", "response": string(respBody), "status": resp.StatusCode})
}

func GetSubscriptions(c fiber.Ctx) error {
	subscriberID, ok := c.Locals("userID").(string)

	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	channelID := c.Params("channelId")

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/subscriptions" +
		"?subscriber_id=eq." + subscriberID +
		"&channel_id=eq." + channelID +
		"&select=*,users!subscriptions_subscriber_id_fkey(id,first_name,last_name,picture)"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch subscriptions: " + string(respBody), "status": resp.StatusCode})
	}
	return c.Status(200).Type("json").Send(respBody)
}

func GetSubscriberCount(c fiber.Ctx) error {
	channelId := c.Params("channelId")

	if channelId == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "channel_id is required",
		})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := fmt.Sprintf(
		"%s/rest/v1/subscriptions?channel_id=eq.%s&select=*",
		supabaseURL,
		channelId,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "count=exact")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return c.Status(500).JSON(fiber.Map{
			"error":   "failed to fetch subscriber count",
			"details": string(body),
		})
	}

	contentRange := resp.Header.Get("Content-Range")

	parts := strings.Split(contentRange, "/")

	count := "0"
	if len(parts) > 1 {
		count = parts[1]
	}

	return c.JSON(fiber.Map{
		"subscriber_count": count,
	})
}
