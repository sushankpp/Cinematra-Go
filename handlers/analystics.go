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

func callRPC(fn string, payload []byte) (int, []byte, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	req, err := http.NewRequest("POST", supabaseURL+"/rest/v1/rpc/"+fn, bytes.NewBuffer(payload))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, nil
}

func rawJSON(b []byte) interface{} {
	var v interface{}
	if len(b) > 0 {
		_ = json.Unmarshal(b, &v)
	}
	return v
}

func GetChannelAnalytics(c fiber.Ctx) error {
	channelID, ok := c.Locals("userID").(string)
	if !ok || channelID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	days := c.Query("days", "30")

	argPayload := []byte(`{"p_channel_id":"` + channelID + `","p_days":` + days + `}`)
	summaryPayload := []byte(`{"p_channel_id":"` + channelID + `"}`)

	_, viewsBody, err := callRPC("get_channel_view_stats", argPayload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "views stats failed: " + err.Error()})
	}
	_, subsBody, err := callRPC("get_channel_subscriber_stats", argPayload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "sub stats failed: " + err.Error()})
	}
	_, reactionsBody, err := callRPC("get_channel_reaction_stats", argPayload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "reaction stats failed: " + err.Error()})
	}
	_, summaryBody, err := callRPC("get_channel_summary", summaryPayload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "summary failed: " + err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{
		"views_over_time":     rawJSON(viewsBody),
		"subs_over_time":      rawJSON(subsBody),
		"reactions_over_time": rawJSON(reactionsBody),
		"summary":             rawJSON(summaryBody),
	})
}
