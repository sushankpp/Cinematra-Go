package recommend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

type trendingRow struct {
	VideoId              string `json:"video_id"`
	WatchDurationSeconds int    `json:"watch_duration_seconds"`
	VideoDurationSeconds int    `json:"video_duration_seconds"`
	Completed            bool   `json:"completed"`
}

// FetchTrendingScore fetches watch history from the last 7 days and computes a trending score for each video.
func FetchTrendingScore() (map[string]float64, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Build the 7-day lookback timestamp in RFC3339 format
	since := time.Now().AddDate(0, 0, -7).UTC().Format(time.RFC3339)
	url := supabaseURL + "/rest/v1/watch_history?select=video_id,watch_duration_seconds,video_duration_seconds,completed&watched_at=gte." + since

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase returned status %s: %s", resp.Status, string(body))
	}

	var rows []trendingRow
	if err = json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}

	trendingScore := make(map[string]float64)
	for _, row := range rows {
		var weight float64
		if row.Completed {
			weight = 1.0
		} else if row.VideoDurationSeconds > 0 {
			weight = float64(row.WatchDurationSeconds) / float64(row.VideoDurationSeconds)
			if weight > 1.0 {
				weight = 1.0
			}
		} else {
			weight = 0.5
		}
		trendingScore[row.VideoId] += weight
	}

	return trendingScore, nil
}

// RankTrending sorts video IDs by their trending score descending and returns the top `limit` IDs.
func RankTrending(scores map[string]float64, limit int) []string {
	type entry struct {
		videoId string
		score   float64
	}

	entries := make([]entry, 0, len(scores))
	for id, score := range scores {
		entries = append(entries, entry{videoId: id, score: score})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score > entries[j].score
	})

	if len(entries) < limit {
		limit = len(entries)
	}

	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, entries[i].videoId)
	}
	return result
}
