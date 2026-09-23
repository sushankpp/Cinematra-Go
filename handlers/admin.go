package handlers

import (
	"bytes"
	"cinematra-backend/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

// createNotification inserts a notification row into the notifications table
// so the affected user can see why their content was removed.
func createNotification(userID, notifType, message string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"type":    notifType,
		"message": message,
	})

	url := supabaseURL + "/rest/v1/notifications"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("createNotification: failed to build request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("createNotification: request failed:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("createNotification: insert failed:", string(body))
	}
}

// ──────────────────────────── Admin Handlers ────────────────────────────

// AdminGetAllUsers returns all users in the system.
func AdminGetAllUsers(c fiber.Ctx) error {
	rows, err := config.DB.Query(context.Background(),
		"SELECT id, first_name, last_name, email, phone, picture, provider, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch users: " + err.Error()})
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, firstName, lastName, email, provider string
		var phone, picture *string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &firstName, &lastName, &email, &phone, &picture, &provider, &createdAt, &updatedAt); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"first_name": firstName,
			"last_name":  lastName,
			"email":      email,
			"phone":      phone,
			"picture":    picture,
			"provider":   provider,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	if users == nil {
		users = []map[string]interface{}{}
	}
	return c.JSON(users)
}

// AdminDeleteUser deletes a user by ID. Cannot delete yourself.
func AdminDeleteUser(c fiber.Ctx) error {
	targetID := c.Params("id")
	callerID := c.Locals("userID").(string)

	if targetID == callerID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot delete yourself"})
	}

	tag, err := config.DB.Exec(context.Background(), "DELETE FROM users WHERE id = $1", targetID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete user: " + err.Error()})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "User deleted successfully"})
}

// AdminGetAllVideos returns all videos (all statuses, all users).
func AdminGetAllVideos(c fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos" +
		"?select=id,title,thumbnail_url,user_id,status,moderation_label,moderation_confidence,created_at,views,users(first_name,last_name,email)" +
		"&order=created_at.desc"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch videos: " + err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(resp.StatusCode).Type("json").Send(body)
}

// AdminGetAllComments returns all comments across all videos.
func AdminGetAllComments(c fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/comments" +
		"?select=id,video_id,user_id,content,moderation_label,moderation_confidence,created_at,users(first_name,last_name,email)" +
		"&order=created_at.desc"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch comments: " + err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(resp.StatusCode).Type("json").Send(body)
}

// GetNotifications returns the logged-in user's notifications.
func GetNotifications(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/notifications?user_id=eq." + userID + "&order=created_at.desc"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch notifications"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(resp.StatusCode).Type("json").Send(body)
}

// MarkNotificationRead marks a single notification as read.
func MarkNotificationRead(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	notifID := c.Params("id")

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload, _ := json.Marshal(map[string]interface{}{"read": true})
	url := supabaseURL + "/rest/v1/notifications?id=eq." + notifID + "&user_id=eq." + userID
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", "application/json")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update notification"})
	}
	defer resp.Body.Close()

	return c.JSON(fiber.Map{"success": true})
}
