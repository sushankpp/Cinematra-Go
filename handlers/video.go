package handlers

import (
	"bytes"
	"cinematra-backend/utils/hls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

func UploadVideo(c fiber.Ctx) error {
	title := c.FormValue("title")
	description := c.FormValue("description")
	if title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "title is required"})
	}

	file, err := c.FormFile("Video")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "video file is required"})
	}

	userID := c.Locals("userID").(string)

	tmpDir := fmt.Sprintf("/tmp/cinematra_%d", time.Now().UnixNano())
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	videoPath := filepath.Join(tmpDir, file.Filename)
	if err := c.SaveFile(file, videoPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save uploaded video"})
	}

	// thumbnail — use user provided or auto generate
	var thumbnailURL string
	userThumb, err := c.FormFile("Thumbnail")
	if err == nil && userThumb != nil {
		thumbPath := filepath.Join(tmpDir, userThumb.Filename)
		if err := c.SaveFile(userThumb, thumbPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save uploaded thumbnail"})
		}
		thumbObjPath := fmt.Sprintf("%s/%d_thumb.jpg", userID, time.Now().UnixNano())
		thumbnailURL, err = uploadToSupabase(thumbPath, thumbObjPath, "Thumbnails", "image/jpeg")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail upload failed: " + err.Error()})
		}
	} else {
		thumbPath, err := generateThumbnail(videoPath, tmpDir)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail generation failed: " + err.Error()})
		}
		info, err := os.Stat(thumbPath)
		if err != nil || info.Size() == 0 {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail generation failed: empty file"})
		}
		thumbObjPath := fmt.Sprintf("%s/%d_thumbnail.jpg", userID, time.Now().UnixNano())
		thumbnailURL, err = uploadToSupabase(thumbPath, thumbObjPath, "Thumbnails", "image/jpeg")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail upload failed: " + err.Error()})
		}
	}

	// build base Supabase URL for HLS segments
	supabaseURL := os.Getenv("SUPABASE_URL")
	baseObjPath := fmt.Sprintf("%s/%d", userID, time.Now().UnixNano())
	baseURL := fmt.Sprintf(
		"%s/storage/v1/object/public/Videos/%s",
		supabaseURL, baseObjPath,
	)

	// generate HLS streams with absolute URLs in master playlist
	hlsDir := filepath.Join(tmpDir, "hls")
	hlsResult, err := hls.GenerateHLS(videoPath, hlsDir, hls.DefaultQualities, baseURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "HLS generation failed: " + err.Error()})
	}

	// upload master playlist
	masterObjPath := fmt.Sprintf("%s/master.m3u8", baseObjPath)
	masterURL, err := uploadToSupabase(hlsResult.MasterPlaylistPath, masterObjPath, "Videos", "application/x-mpegURL")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "master playlist upload failed: " + err.Error()})
	}

	// upload all quality segments and playlists
	for qualityName, qualityDir := range hlsResult.QualityDirs {
		entries, _ := filepath.Glob(filepath.Join(qualityDir, "*"))
		for _, entry := range entries {
			filename := filepath.Base(entry)
			objPath := fmt.Sprintf("%s/%s/%s", baseObjPath, qualityName, filename)
			contentType := "application/x-mpegURL"
			if filepath.Ext(filename) == ".ts" {
				contentType = "video/MP2T"
			}
			if _, err := uploadToSupabase(entry, objPath, "Videos", contentType); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("chunk upload failed for %s/%s: %s", qualityName, filename, err.Error()),
				})
			}
		}
	}

	// save to DB
	videoID, err := saveVideoMetadata(userID, title, description, masterURL, thumbnailURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save video metadata: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"video_id":      videoID,
		"title":         title,
		"description":   description,
		"master_url":    masterURL,
		"thumbnail_url": thumbnailURL,
		"qualities":     []string{"1080p", "720p", "480p", "360p"},
		"status":        "ready",
	})
}

func UpdateVideo(c fiber.Ctx) error {
	videoID := c.Params("id")
	userID := c.Locals("userID").(string)

	title := c.FormValue("title")
	description := c.FormValue("description")

	payload := "{"
	if title != "" {
		payload += fmt.Sprintf(`"title":%q,`, title)
	}
	if description != "" {
		payload += fmt.Sprintf(`"description":%q,`, description)
	}

	userThumb, err := c.FormFile("thumbnail")
	if err == nil && userThumb != nil {
		tmpDir := fmt.Sprintf("/tmp/cinematra_thumb_%d", time.Now().UnixNano())
		os.MkdirAll(tmpDir, 0755)
		defer os.RemoveAll(tmpDir)

		thumbPath := filepath.Join(tmpDir, userThumb.Filename)
		if err := c.SaveFile(userThumb, thumbPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save thumbnail"})
		}
		thumbObjPath := fmt.Sprintf("%s/%d_thumb.jpg", userID, time.Now().UnixNano())
		thumbnailURL, err := uploadToSupabase(thumbPath, thumbObjPath, "Thumbnails", "image/jpeg")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail upload failed: " + err.Error()})
		}
		payload += fmt.Sprintf(`"thumbnail_url":%q,`, thumbnailURL)
	}

	if payload == "{" {
		return c.Status(400).JSON(fiber.Map{"error": "nothing to update"})
	}

	payload = payload[:len(payload)-1] + "}"

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := fmt.Sprintf("%s/rest/v1/videos?id=eq.%s&user_id=eq.%s", supabaseURL, videoID, userID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBufferString(payload))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "update failed: " + string(body)})
	}

	return c.JSON(fiber.Map{"message": "video updated successfully"})
}

func generateThumbnail(videoPath, outDir string) (string, error) {
	thumbPath := filepath.Join(outDir, "thumbnail.jpg")
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:02.000", "-vframes", "1", "-vf", "scale=1280:720", "-q:v", "3", thumbPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		cmd2 := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:00", "-vframes", "1", "-vf", "scale=1280:720", "-q:v", "3", thumbPath)
		cmd2.Stderr = &stderr
		if err2 := cmd2.Run(); err2 != nil {
			return "", fmt.Errorf("ffmpeg thumbnail: %s", stderr.String())
		}
	}
	return thumbPath, nil
}

func uploadToSupabase(filePath, objectPath, bucket, contentType string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, objectPath)
	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(body))
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, objectPath), nil
}

func saveVideoMetadata(userID, title, description, masterURL, thumbnailURL string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload := fmt.Sprintf(
		`{"user_id":%q,"title":%q,"description":%q,"storage_url":%q,"thumbnail_url":%q,"status":"ready"}`,
		userID, title, description, masterURL, thumbnailURL,
	)

	url := supabaseURL + "/rest/v1/videos"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("DB insert failed: %s", string(body))
	}

	return extractField(string(body), "id"), nil
}

func GetVideos(c fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos?status=eq.ready&select=*,users(first_name,last_name,picture)&order=created_at.desc"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch videos: " + err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(200).Type("json").Send(body)
}

func GetVideo(c fiber.Ctx) error {
	id := c.Params("id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos?id=eq." + id + "&select=*,users(first_name,last_name,picture)"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Accept", "application/vnd.pgrst.object+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch video: " + err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return c.Status(404).JSON(fiber.Map{"error": "video not found"})
	}
	body, _ := io.ReadAll(resp.Body)
	return c.Status(200).Type("json").Send(body)
}

func StreamInfo(c fiber.Ctx) error {
	bufferLevel, _ := strconv.ParseFloat(c.Query("buffer", "0"), 64)

	state := hls.NewBOLAState(hls.DefaultQualities, 1.0, 24.0)
	decision := state.NextQuality(bufferLevel)

	return c.JSON(fiber.Map{
		"recommended_quality": decision.Quality.Name,
		"quality_index":       decision.QualityIndex,
		"bitrate":             decision.Quality.Bitrate,
		"reason": fmt.Sprintf(
			"buffer=%.1fs → picked %s (score=%.4f)",
			bufferLevel, decision.Quality.Name, decision.Score,
		),
	})
}

func extractField(body, key string) string {
	search := fmt.Sprintf(`"%s":"`, key)
	idx := findIndex(body, search)
	if idx == -1 {
		return ""
	}
	start := idx + len(search)
	end := findIndex(body[start:], `"`)
	if end == -1 {
		return ""
	}
	return body[start : start+end]
}

func findIndex(s, substr string) int {
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
