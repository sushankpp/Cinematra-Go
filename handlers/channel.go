package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

func GetChannels(c fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// 1. Fetch all channels with precomputed stats
	url := supabaseURL + "/rest/v1/channel_stats?video_count=gt.0&order=subscriber_count.desc"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var channels []map[string]interface{}
	if err := json.Unmarshal(body, &channels); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to parse channels"})
	}

	// 2. If logged in, fetch the viewer's subscribed channel IDs in ONE query
	subscriberID, ok := c.Locals("userID").(string)
	if ok && subscriberID != "" {
		subURL := supabaseURL + "/rest/v1/subscriptions?subscriber_id=eq." + subscriberID + "&select=channel_id"
		subReq, _ := http.NewRequest("GET", subURL, nil)
		subReq.Header.Set("apikey", serviceKey)
		subReq.Header.Set("Authorization", "Bearer "+serviceKey)

		subResp, err := httpClient.Do(subReq)
		if err == nil {
			defer subResp.Body.Close()
			subBody, _ := io.ReadAll(subResp.Body)

			var subs []map[string]interface{}
			json.Unmarshal(subBody, &subs)

			subscribedSet := make(map[string]bool, len(subs))
			for _, s := range subs {
				if chID, ok := s["channel_id"].(string); ok {
					subscribedSet[chID] = true
				}
			}

			// 3. Merge is_subscribed flag into each channel
			for i, ch := range channels {
				if chID, ok := ch["channel_id"].(string); ok {
					channels[i]["is_subscribed"] = subscribedSet[chID]
				}
			}
		}
	} else {
		for i := range channels {
			channels[i]["is_subscribed"] = false
		}
	}

	return c.Status(200).JSON(channels)
}
