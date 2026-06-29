package recommend

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

type watchRow struct {
	VideoID              string `json:"video_id"`
	WatchDurationSeconds int    `json:"watch_duration_seconds"`
	VideoDurationSeconds int    `json:"video_duration_seconds"`
	Completed            bool   `json:"completed"`
}

//this file fetches users watch history and merges those video vector into a single "user" vector

func FetchWatchHistory(userId string) ([]WatchedVideo, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := supabaseURL + "/rest/v1/watch_history?user_id=eq." + userId + "&select=video_id,watch_duration_seconds,video_duration_seconds,completed"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rows []watchRow
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("failed to parse video rows: %w", err)
	}

	var watchedVideos []WatchedVideo
	for _, row := range rows {

		if row.Completed == true {
			watchedVideos = append(watchedVideos, WatchedVideo{
				VideoId: row.VideoID,
				Weight:  1.0,
			})
		} else if row.VideoDurationSeconds > 0 {
			watchedVideos = append(watchedVideos, WatchedVideo{
				VideoId: row.VideoID,
				Weight:  float64(row.WatchDurationSeconds) / float64(row.VideoDurationSeconds),
			})
		} else {
			watchedVideos = append(watchedVideos, WatchedVideo{
				VideoId: row.VideoID,
				Weight:  0.5,
			})
		}
	}

	return watchedVideos, nil
}

func BuildUserProfile(watched []WatchedVideo, corpus *Corpus) map[string]float64 {
	profile := make(map[string]float64)

	for _, rows := range watched {
		video, exists := corpus.VideoVectors[rows.VideoId]
		if !exists {
			continue
		}
		for word, score := range video.Vector {
			profile[word] += score * rows.Weight
		}
		//normalize the profile vector
		var magnitude float64
		for _, score := range profile {
			magnitude += score * score
		}
		magnitude = math.Sqrt(magnitude)
		if magnitude == 0 {
			continue
		}
		for word := range profile {
			profile[word] /= magnitude
		}
	}

	return profile
}
