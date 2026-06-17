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
	language := c.FormValue("language")
	contentRating := c.FormValue("content_rating")
	releaseYearStr := c.FormValue("release_year")

	var releaseYear *int
	if releaseYearStr != "" {
		if y, err := strconv.Atoi(releaseYearStr); err == nil {
			releaseYear = &y
		}
	}

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

	supabaseURL := os.Getenv("SUPABASE_URL")
	baseObjPath := fmt.Sprintf("%s/%d", userID, time.Now().UnixNano())
	baseURL := fmt.Sprintf("%s/storage/v1/object/public/Videos/%s", supabaseURL, baseObjPath)

	hlsDir := filepath.Join(tmpDir, "hls")
	hlsResult, err := hls.GenerateHLS(videoPath, hlsDir, hls.DefaultQualities, baseURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "HLS generation failed: " + err.Error()})
	}

	masterObjPath := fmt.Sprintf("%s/master.m3u8", baseObjPath)
	masterURL, err := uploadToSupabase(hlsResult.MasterPlaylistPath, masterObjPath, "Videos", "application/x-mpegURL")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "master playlist upload failed: " + err.Error()})
	}

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

	// Save to DB first so we have videoID
	videoID, err := saveVideoMetadata(userID, title, description, masterURL, thumbnailURL, language, contentRating, releaseYear)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save video metadata: " + err.Error()})
	}

	// Now genres and tags can use the valid videoID
	genreIDs := c.FormValue("genre_ids")
	if genreIDs != "" {
		if err := saveVideoGenres(videoID, genreIDs); err != nil {
			fmt.Println("failed to save genres:", err)
		}
	}

	tagIDs := c.FormValue("tag_ids")
	if tagIDs != "" {
		if err := saveVideoTags(videoID, tagIDs); err != nil {
			fmt.Println("failed to save tags:", err)
		}
	}

	fmt.Println(">>> video uploaded successfully with ID:", videoID)

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

	fields := map[string]string{
		"title":          c.FormValue("title"),
		"description":    c.FormValue("description"),
		"language":       c.FormValue("language"),
		"content_rating": c.FormValue("content_rating"),
	}

	updates := map[string]interface{}{}
	for col, val := range fields {
		if val != "" {
			updates[col] = val
		}
	}

	if yrStr := c.FormValue("release_year"); yrStr != "" {
		if y, err := strconv.Atoi(yrStr); err == nil {
			updates["release_year"] = y
		}
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
		updates["thumbnail_url"] = thumbnailURL
	}

	genreIDs := c.FormValue("genre_ids")
	if genreIDs != "" {
		if err := replaceVideoGenres(videoID, genreIDs); err != nil {
			fmt.Println("failed to update genres:", err)
		}
	}

	tagIDs := c.FormValue("tag_ids")
	if tagIDs != "" {
		if err := replaceVideoTags(videoID, tagIDs); err != nil {
			fmt.Println("failed to update tags:", err)
		}
	}

	if len(updates) > 0 {
		if err := patchVideoRow(videoID, userID, updates); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	} else if genreIDs == "" && tagIDs == "" {
		return c.Status(400).JSON(fiber.Map{"error": "nothing to update"})
	}

	return c.JSON(fiber.Map{"message": "video updated successfully"})
}

func DeleteVideo(c fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	videoId := c.Params("id")

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos?id=eq." + videoId + "&user_id=eq." + userID
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create delete request: " + err.Error()})
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete video: " + err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "delete failed: " + string(body)})
	}
	return c.Status(200).JSON(fiber.Map{"success": true, "message": "Video deleted successfully"})
}

func patchVideoRow(videoID, userID string, updates map[string]interface{}) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	parts := []string{}
	for k, v := range updates {
		switch val := v.(type) {
		case string:
			parts = append(parts, fmt.Sprintf(`"%s":%q`, k, val))
		case int:
			parts = append(parts, fmt.Sprintf(`"%s":%d`, k, val))
		}
	}
	payload := "{" + joinStrings(parts, ",") + "}"

	url := fmt.Sprintf("%s/rest/v1/videos?id=eq.%s&user_id=eq.%s", supabaseURL, videoID, userID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed: %s", string(body))
	}
	return nil
}

func replaceVideoGenres(videoID, genreIDs string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	delURL := fmt.Sprintf("%s/rest/v1/video_genres?video_id=eq.%s", supabaseURL, videoID)
	delReq, err := http.NewRequest("DELETE", delURL, nil)
	if err != nil {
		return err
	}
	delReq.Header.Set("Authorization", "Bearer "+serviceKey)
	delReq.Header.Set("apikey", serviceKey)
	resp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return saveVideoGenres(videoID, genreIDs)
}

func replaceVideoTags(videoID, tagIDs string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	delURL := fmt.Sprintf("%s/rest/v1/video_tags?video_id=eq.%s", supabaseURL, videoID)
	delReq, err := http.NewRequest("DELETE", delURL, nil)
	if err != nil {
		return err
	}
	delReq.Header.Set("Authorization", "Bearer "+serviceKey)
	delReq.Header.Set("apikey", serviceKey)
	resp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return saveVideoTags(videoID, tagIDs)
}

func IncrementVideoViews(c fiber.Ctx) error {
	id := c.Params("id")
	fmt.Println(">>> increment views called for id:", id)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	payload := fmt.Sprintf(`{"video_id":%q}`, id)
	url := supabaseURL + "/rest/v1/rpc/increment_video_views"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payload))
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

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{"error": "view increment failed: " + string(body)})
	}

	return c.JSON(fiber.Map{"message": "view counted"})
}

func generateThumbnail(videoPath, outDir string) (string, error) {
	thumbPath := filepath.Join(outDir, "thumbnail.jpg")
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:02.000", "-vframes", "1", "-vf", "scale=640:360", "-q:v", "3", thumbPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		cmd2 := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:00", "-vframes", "1", "-vf", "scale=640:360", "-q:v", "3", thumbPath)
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

func saveVideoMetadata(userID, title, description, masterURL, thumbnailURL, language, contentRating string, releaseYear *int) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	releaseYearVal := "null"
	if releaseYear != nil {
		releaseYearVal = strconv.Itoa(*releaseYear)
	}

	payload := fmt.Sprintf(
		`{"user_id":%q,"title":%q,"description":%q,"storage_url":%q,"thumbnail_url":%q,"status":"ready","language":%q,"content_rating":%q,"release_year":%s}`,
		userID, title, description, masterURL, thumbnailURL, language, contentRating, releaseYearVal,
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

	url := supabaseURL + "/rest/v1/videos" +
		"?status=eq.ready" +
		"&select=id,title,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))" +
		"&order=created_at.desc" +
		"&limit=20"

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

	url := supabaseURL + "/rest/v1/videos" +
		"?id=eq." + id +
		"&select=id,title,description,storage_url,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))"

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

func GetUserVideo(c fiber.Ctx) error {
	id := c.Locals("userID").(string)
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos" + "?user_id=eq." + id + "&select=id,title,description,thumbnail_url,user_id,created_at,release_year,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error ": "failed to fetch video: " + err.Error()})
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

func saveVideoGenres(videoID, genreIDs string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	var rows []string
	for _, id := range splitIDs(genreIDs) {
		rows = append(rows, fmt.Sprintf(`{"video_id":%q,"genre_id":%s}`, videoID, id))
	}
	payload := "[" + joinStrings(rows, ",") + "]"

	url := supabaseURL + "/rest/v1/video_genres"
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(payload))
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("genre insert failed: %s", string(body))
	}
	return nil
}

func saveVideoTags(videoID, tagIDs string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	var rows []string
	for _, id := range splitIDs(tagIDs) {
		rows = append(rows, fmt.Sprintf(`{"video_id":%q,"tag_id":%s}`, videoID, id))
	}
	payload := "[" + joinStrings(rows, ",") + "]"

	url := supabaseURL + "/rest/v1/video_tags"
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(payload))
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tag insert failed: %s", string(body))
	}
	return nil
}

func splitIDs(s string) []string {
	var result []string
	for _, part := range splitString(s, ",") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
