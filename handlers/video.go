package handlers

import (
	"bytes"
	"cinematra-backend/utils/compression"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
)

func UploadVideo(c fiber.Ctx) error {
	// Parse the data from the form fields
	title := c.FormValue("title")
	description := c.FormValue("description")
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}

	// get the uploaded file i.e video
	file, err := c.FormFile("Video")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "video file is required"})
	}

	// get userID from context (set by auth middleware)
	userID := c.Locals("userID").(string)

	// save the uploaded file to a temporary location
	tmpDir := fmt.Sprintf("/tmp/cinematra_%d", time.Now().UnixNano())
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir) // clean up temp files

	videoPath := filepath.Join(tmpDir, file.Filename)
	if err := c.SaveFile(file, videoPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save uploaded video"})
	}

	//upload raw video to supabase storage
	videoObjPath := fmt.Sprintf("%s/%d_%s", userID, time.Now().UnixNano(), file.Filename)
	storageURL, err := uploadToSupabase(videoPath, videoObjPath, "Videos", "video/mp4")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "video upload failed: " + err.Error()})
	}

	//generate thumbnail
	thumbPath, err := generateThumbnail(videoPath, tmpDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "thumbnail generation failed: " + err.Error()})
	}

	//check thumbnail file exists
	info, err := os.Stat(thumbPath)
	if err != nil || info.Size() == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "thumbnail generation produced empty file"})
	}

	//upload thumbnail to supabase storage
	thumbObjPath := fmt.Sprintf("%s/%d_thumb.jpg", userID, time.Now().UnixNano())
	thumbnailURL, err := uploadToSupabase(thumbPath, thumbObjPath, "Thumbnails", "image/jpeg")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "thumbnail upload failed: " + err.Error()})
	}

	//extract frames and run compression pipeline
	framesDir := filepath.Join(tmpDir, "frames")
	os.MkdirAll(framesDir, 0755)
	if err := extractFrames(videoPath, framesDir); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "frame extraction failed: " + err.Error()})
	}
	stats, err := compressFrames(framesDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "frame compression failed: " + err.Error()})
	}

	//save metadata to DB
	videoID, err := saveVideoMetadata(userID, title, description, storageURL, thumbnailURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save video metadata: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"video_id":          videoID,
		"title":             title,
		"description":       description,
		"video_url":         storageURL,
		"thumbnail_url":     thumbnailURL,
		"compression_stats": stats,
	})
}

// generate thumbnail using ffmpeg by extracting the single frame at 2sec and saving it as a jpeg
func generateThumbnail(videoPath, outDir string) (string, error) {
	thumbPath := filepath.Join(outDir, "thumbnail.jpg")

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:02.000", "-vframes", "1", "-vf", "scale=1280:720", "-q:v", "3", thumbPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		cmd2 := exec.Command(
			"ffmpeg",
			"-i", videoPath,
			"-ss", "00:00:00",
			"-vframes", "1",
			"-vf", "scale=1280:720",
			"-q:v", "3",
			thumbPath,
		)
		cmd2.Stderr = &stderr
		if err2 := cmd2.Run(); err2 != nil {
			return "", fmt.Errorf("ffmpeg: %s", stderr.String())
		}
	}
	return thumbPath, nil
}

// extractframes using fmmpeg to dump 1 PNG frame per second
func extractFrames(videoPath, outDir string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1", "-q:v", "2", filepath.Join(outDir, "frame_%04d.png"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %s", stderr.String())
	}
	return nil
}

// compress frames uses the DCT + frame diff pipline on extracted frames
func compressFrames(framesDir string) (fiber.Map, error) {
	entries, err := filepath.Glob(filepath.Join(framesDir, "*.png"))
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no frames found for compression")
	}

	encoder := compression.NewVideoEncoder(75, 30)

	for _, framePath := range entries {
		f, err := os.Open(framePath)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		encoder.AddFrame(img)
	}

	stats := encoder.Stats()

	return fiber.Map{
		"frame_count":    len(entries),
		"i_frames":       stats["i_frames"],
		"p_frames":       stats["p_frames"],
		"skipped_blocks": stats["skipped_blocks"],
		"encoded_blocks": stats["encoded_blocks"],
	}, nil
}

// uploadToSupbase function uploads any file to a supbase storage bucket
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

	publicUrl := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, objectPath)
	return publicUrl, nil
}

// saveVideoMetadata saves the video into DB
func saveVideoMetadata(userID, title, description, videoURL, thumbnailURL string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload := fmt.Sprintf(`{"user_id":%q,"title":%q,"description":%q,"storage_url":%q,"thumbnail_url":%q,"status":"ready"}`, userID, title, description, videoURL, thumbnailURL)

	url := supabaseURL + "/rest/v1/videos"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	req.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("DB insert failed: %s", string(body))
	}

	id := extractField(string(body), "id")

	return id, nil
}

// extractField pulls a string value from a flat JSON string
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

	url := supabaseURL + "/rest/v1/videos?id=eq." + id + "&select=*"
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
