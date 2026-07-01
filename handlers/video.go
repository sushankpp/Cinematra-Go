package handlers

import (
	"bytes"
	"cinematra-backend/utils/hls"
	"encoding/json"
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

func saveToLocal(filePath, objectPath, folder string) (string, error) {
	destDir := filepath.Join("./storage", folder, filepath.Dir(objectPath))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	dest := filepath.Join("./storage", folder, objectPath)
	src, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	host := os.Getenv("SERVER_HOST")
	return fmt.Sprintf("%s/media/%s/%s", host, folder, objectPath), nil
}

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

	// ── Thumbnail ──────────────────────────────────────────────
	baseObjPath := fmt.Sprintf("%s/%d", userID, time.Now().UnixNano())
	var thumbnailURL string

	userThumb, err := c.FormFile("Thumbnail")
	if err == nil && userThumb != nil {
		thumbPath := filepath.Join(tmpDir, userThumb.Filename)
		if err := c.SaveFile(userThumb, thumbPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save uploaded thumbnail"})
		}
		thumbObjPath := fmt.Sprintf("%s/%d_thumb.jpg", userID, time.Now().UnixNano())
		thumbnailURL, err = saveToLocal(thumbPath, thumbObjPath, "Thumbnails")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail save failed: " + err.Error()})
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
		thumbnailURL, err = saveToLocal(thumbPath, thumbObjPath, "Thumbnails")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail save failed: " + err.Error()})
		}
	}

	// ── HLS Generation ─────────────────────────────────────────
	host := os.Getenv("SERVER_HOST")
	baseURL := fmt.Sprintf("%s/media/Videos/%s", host, baseObjPath)

	hlsDir := filepath.Join(tmpDir, "hls")
	hlsResult, err := hls.GenerateHLS(videoPath, hlsDir, hls.DefaultQualities, baseURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "HLS generation failed: " + err.Error()})
	}

	// ── Save master.m3u8 ───────────────────────────────────────
	masterObjPath := fmt.Sprintf("%s/master.m3u8", baseObjPath)
	masterURL, err := saveToLocal(hlsResult.MasterPlaylistPath, masterObjPath, "Videos")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "master playlist save failed: " + err.Error()})
	}

	// ── Save all chunks ────────────────────────────────────────
	for qualityName, qualityDir := range hlsResult.QualityDirs {
		entries, _ := filepath.Glob(filepath.Join(qualityDir, "*"))
		for _, entry := range entries {
			filename := filepath.Base(entry)
			objPath := fmt.Sprintf("%s/%s/%s", baseObjPath, qualityName, filename)
			if _, err := saveToLocal(entry, objPath, "Videos"); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("chunk save failed for %s/%s: %s", qualityName, filename, err.Error()),
				})
			}
		}
	}

	// ── Save metadata to Supabase DB ───────────────────────────
	videoID, err := saveVideoMetadata(userID, title, description, masterURL, thumbnailURL, language, contentRating, releaseYear)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save video metadata: " + err.Error()})
	}

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
		thumbnailURL, err := saveToLocal(thumbPath, thumbObjPath, "Thumbnails")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "thumbnail save failed: " + err.Error()})
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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
	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(delReq)
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
	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(delReq)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch videos: " + err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var videos []map[string]interface{}
	if err := json.Unmarshal(body, &videos); err != nil {
		return c.Status(200).Type("json").Send(body)
	}

	//if user is logged in fetch user's watch progress
	userID, ok := c.Locals("userID").(string)
	fmt.Println("User Id: ", userID, "ok: ", ok)
	if ok && userID != "" {
		histURL := supabaseURL + "/rest/v1/watch_history" +
			"?user_id=eq." + userID +
			"&select=video_id,watch_duration_seconds,video_duration_seconds,completed,watched_at"

		hReq, _ := http.NewRequest("GET", histURL, nil)
		hReq.Header.Set("apikey", serviceKey)
		hReq.Header.Set("Authorization", "Bearer "+serviceKey)

		hResp, err := httpClient.Do(hReq)
		if err == nil {
			defer hResp.Body.Close()
			hBody, _ := io.ReadAll(hResp.Body)

			var history []map[string]interface{}
			json.Unmarshal(hBody, &history)

			// build lookup map: video_id → history row
			progressMap := make(map[string]map[string]interface{})
			for _, h := range history {
				if vid, ok := h["video_id"].(string); ok {
					progressMap[vid] = h
				}
			}

			// merge progress into each video
			for i, v := range videos {
				if vid, ok := v["id"].(string); ok {
					if h, exists := progressMap[vid]; exists {
						videos[i]["watch_duration_seconds"] = h["watch_duration_seconds"]
						videos[i]["video_duration_seconds"] = h["video_duration_seconds"]
						videos[i]["completed"] = h["completed"]
						videos[i]["last_watched"] = h["watched_at"]
					} else {
						videos[i]["watch_duration_seconds"] = nil
						videos[i]["video_duration_seconds"] = nil
						videos[i]["completed"] = nil
						videos[i]["last_watched"] = nil
					}
				}
			}
		}
	}
	return c.Status(200).JSON(videos)
}

func GetVideo(c fiber.Ctx) error {
	id := c.Params("id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	url := supabaseURL + "/rest/v1/videos" +
		"?id=eq." + id +
		"&select=id,title,description,storage_url,thumbnail_url,user_id,release_year,created_at,views,likes_count,dislikes_count,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Accept", "application/vnd.pgrst.object+json")

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch video: " + err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return c.Status(404).JSON(fiber.Map{"error": "video not found"})
	}
	body, _ := io.ReadAll(resp.Body)

	var video map[string]interface{}
	if err := json.Unmarshal(body, &video); err != nil {
		return c.Status(200).Type("json").Send(body)
	}

	// merge logged-in user's reaction, if any
	userID, ok := c.Locals("userID").(string)
	fmt.Println("[DEBUG] GetVideo id:", id, "| userID:", userID, "| ok:", ok, "| Cookie token:", c.Cookies("token"), "| Auth header:", c.Get("Authorization"))
	video["user_reaction"] = nil
	if ok && userID != "" {
		reactURL := supabaseURL + "/rest/v1/video_reactions" +
			"?video_id=eq." + id +
			"&user_id=eq." + userID +
			"&select=reaction_type"

		fmt.Println("GetVideo reaction check URL:", reactURL)

		rReq, _ := http.NewRequest("GET", reactURL, nil)
		rReq.Header.Set("apikey", serviceKey)
		rReq.Header.Set("Authorization", "Bearer "+serviceKey)
		rReq.Header.Set("Accept", "application/vnd.pgrst.object+json")

		rResp, err := httpClient.Do(rReq)
		if err == nil {
			defer rResp.Body.Close()
			if rResp.StatusCode == 200 {
				rBody, _ := io.ReadAll(rResp.Body)
				var reaction struct {
					ReactionType string `json:"reaction_type"`
				}
				json.Unmarshal(rBody, &reaction)
				video["user_reaction"] = reaction.ReactionType
			}
		} else {
			fmt.Println("GetVideo reaction check failed:", err)
		}
	}

	return c.Status(200).JSON(video)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
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
