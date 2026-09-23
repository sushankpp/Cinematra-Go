package handlers

import (
	"bytes"
	"cinematra-backend/utils/moderation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gofiber/fiber/v3"
)

func PostComment(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var body struct {
		VideoID string `json:"video_id"`
		Content string `json:"content"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.VideoID == "" || body.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "video_id and content are required"})
	}

	result := moderation.Classify(moderation.GlobalModel, body.Content)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload, err := json.Marshal(map[string]interface{}{
		"user_id":               userID,
		"video_id":              body.VideoID,
		"content":               body.Content,
		"moderation_label":      result.Label,
		"moderation_confidence": result.Confidence,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create payload"})
	}

	url := supabaseURL + "/rest/v1/comments"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to post comment"})
	}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to post comment: %s", string(respBody)), "status": resp.StatusCode})

	}

	return c.Status(201).JSON(fiber.Map{"message": "comment posted successfully"})
}

func GetComments(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "video id is required"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := supabaseURL + "/rest/v1/comments?video_id=eq." + id + "&select=*,users(first_name,last_name,picture)&order=created_at.desc"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch comments"})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var comments []map[string]interface{}
	if err := json.Unmarshal(body, &comments); err != nil {
		return c.Status(200).Type("json").Send(body)
	}

	labelRank := map[string]int{
		"Safe":                   0,
		"PotentiallyProblematic": 1,
		"Problematic":            2,
	}

	sort.SliceStable(comments, func(i, j int) bool {
		labelI, _ := comments[i]["moderation_label"].(string)
		labelJ, _ := comments[j]["moderation_label"].(string)
		return labelRank[labelI] < labelRank[labelJ]
	})

	return c.Status(200).JSON(comments)
}

func DeleteComment(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	commentID := c.Params("id")
	role, _ := c.Locals("userRole").(string)

	if commentID == "" || userID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Comment ID and user ID are required"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// If superadmin, find comment owner and content first for notification
	var ownerID string
	var commentContent string
	if role == "superadmin" {
		ownerURL := supabaseURL + "/rest/v1/comments?id=eq." + commentID + "&select=user_id,content"
		oReq, _ := http.NewRequest("GET", ownerURL, nil)
		oReq.Header.Set("apikey", serviceKey)
		oReq.Header.Set("Authorization", "Bearer "+serviceKey)
		oReq.Header.Set("Accept", "application/vnd.pgrst.object+json")

		var httpClient = &http.Client{Timeout: 10 * time.Second}
		oResp, err := httpClient.Do(oReq)
		if err == nil {
			defer oResp.Body.Close()
			oBody, _ := io.ReadAll(oResp.Body)
			var commentRow struct {
				UserID  string `json:"user_id"`
				Content string `json:"content"`
			}
			json.Unmarshal(oBody, &commentRow)
			ownerID = commentRow.UserID
			commentContent = commentRow.Content
		}
	}

	// Build delete URL — superadmin skips user_id filter
	url := supabaseURL + "/rest/v1/comments?id=eq." + commentID
	if role != "superadmin" {
		url += "&user_id=eq." + userID
	}
	url += "&select=*"

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete comment"})
	}
	defer resp.Body.Close()

	// If superadmin deleted someone else's comment, create a notification
	if role == "superadmin" && ownerID != "" && ownerID != userID {
		reason := c.Query("reason")
		preview := commentContent
		runes := []rune(preview)
		if len(runes) > 40 {
			preview = string(runes[:40]) + "..."
		}

		var message string
		if preview != "" {
			if reason != "" {
				message = fmt.Sprintf("Your comment \"%s\" was removed by an administrator. Reason: %s", preview, reason)
			} else {
				message = fmt.Sprintf("Your comment \"%s\" was removed by an administrator.", preview)
			}
		} else {
			if reason != "" {
				message = fmt.Sprintf("Your comment was removed by an administrator. Reason: %s", reason)
			} else {
				message = "Your comment was removed by an administrator."
			}
		}
		createNotification(ownerID, "comment_deleted", message)
	}

	body, _ := io.ReadAll(resp.Body)
	return c.Status(200).Type("json").Send(body)
}
