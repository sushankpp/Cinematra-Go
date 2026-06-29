package handlers

import (
	"cinematra-backend/utils/recommend"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// GetRecommendations handles recommendations for the current user
func GetRecommendations(c fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Pull precomputed corpus + trending data from cache (refreshed in background)
	// instead of rebuilding the entire TF-IDF corpus on every request.
	corpus, trendingScores, maxTrending := recommend.Cache.Snapshot()
	if corpus == nil {
		return c.Status(503).JSON(fiber.Map{"error": "recommendations warming up, try again shortly"})
	}

	// Fetch user's watch history
	watched, err := recommend.FetchWatchHistory(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch watch history"})
	}

	normalizedTrend := func(videoID string) float64 {
		if maxTrending <= 0 {
			return 0.0
		}
		if s, exists := trendingScores[videoID]; exists {
			return s / maxTrending
		}
		return 0.0
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Helper function to fetch full video details from Supabase using a given URL
	fetchVideosFromURL := func(url string) ([]byte, error) {
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

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("database returned status %s: %s", resp.Status, string(body))
		}
		return body, nil
	}

	// Cold start: User has no watch history — return latest videos with
	// similarity_score: 0, but still attach trending_score / combined_score.
	if len(watched) == 0 {
		url := supabaseURL + "/rest/v1/videos" +
			"?status=eq.ready" +
			"&select=id,title,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))" +
			"&order=created_at.desc" +
			"&limit=10"

		body, err := fetchVideosFromURL(url)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch default recommendations: " + err.Error()})
		}

		var coldStartVideos []map[string]interface{}
		if err := json.Unmarshal(body, &coldStartVideos); err != nil {
			return c.Status(200).Type("json").Send(body)
		}
		for i := range coldStartVideos {
			ts := 0.0
			if id, ok := coldStartVideos[i]["id"].(string); ok {
				ts = normalizedTrend(id)
			}
			coldStartVideos[i]["similarity_score"] = 0.0
			coldStartVideos[i]["trending_score"] = math.Round(ts*10000) / 10000
			coldStartVideos[i]["combined_score"] = math.Round(0.3*ts*10000) / 10000
		}
		return c.Status(200).JSON(coldStartVideos)
	}

	// Create map of watched video IDs to exclude them from recommendation candidates
	watchedIDs := make(map[string]bool)
	for _, w := range watched {
		watchedIDs[w.VideoId] = true
	}

	// Build user preference profile and generate recommended video IDs
	userProfile := recommend.BuildUserProfile(watched, corpus)
	scored := recommend.Recommend(userProfile, corpus, watchedIDs, 10)

	// If no matching unwatched videos, fall back to latest videos but still compute scores
	if len(scored) == 0 {
		url := supabaseURL + "/rest/v1/videos" +
			"?status=eq.ready" +
			"&select=id,title,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))" +
			"&order=created_at.desc" +
			"&limit=10"

		body, err := fetchVideosFromURL(url)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch fallback recommendations: " + err.Error()})
		}

		var fallbackVideos []map[string]interface{}
		if err := json.Unmarshal(body, &fallbackVideos); err != nil {
			return c.Status(200).Type("json").Send(body)
		}

		// Attach similarity, trending, and combined scores to fallback videos
		for i, v := range fallbackVideos {
			if id, ok := v["id"].(string); ok {
				sim := 0.0
				if vec, exists := corpus.VideoVectors[id]; exists {
					sim = recommend.CosineSimilarity(userProfile, vec.Vector)
				}
				ts := normalizedTrend(id)
				fallbackVideos[i]["similarity_score"] = math.Round(sim*10000) / 10000
				fallbackVideos[i]["trending_score"] = math.Round(ts*10000) / 10000
				fallbackVideos[i]["combined_score"] = math.Round((0.7*sim+0.3*ts)*10000) / 10000
			}
		}

		return c.Status(200).JSON(fallbackVideos)
	}

	// Build ID list and score lookup map
	recommendedIDs := make([]string, 0, len(scored))
	scoreMap := make(map[string]float64, len(scored))
	for _, sv := range scored {
		recommendedIDs = append(recommendedIDs, sv.VideoId)
		scoreMap[sv.VideoId] = sv.Similarity
	}

	// Fetch full video details for recommended IDs
	url := supabaseURL + "/rest/v1/videos" +
		"?id=in.(" + strings.Join(recommendedIDs, ",") + ")" +
		"&select=id,title,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))"

	body, err := fetchVideosFromURL(url)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch recommended videos details: " + err.Error()})
	}

	var videos []map[string]interface{}
	if err := json.Unmarshal(body, &videos); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse recommended videos"})
	}

	// Reorder to match the recommendation similarity descending sort order
	// and attach similarity_score / trending_score / combined_score to each video
	videoMap := make(map[string]map[string]interface{})
	for _, v := range videos {
		if id, ok := v["id"].(string); ok {
			videoMap[id] = v
		}
	}

	orderedVideos := make([]map[string]interface{}, 0, len(recommendedIDs))
	for _, id := range recommendedIDs {
		if v, exists := videoMap[id]; exists {
			sim := scoreMap[id]
			ts := normalizedTrend(id)
			v["similarity_score"] = math.Round(sim*10000) / 10000
			v["trending_score"] = math.Round(ts*10000) / 10000
			v["combined_score"] = math.Round((0.7*sim+0.3*ts)*10000) / 10000
			orderedVideos = append(orderedVideos, v)
		}
	}

	return c.Status(200).JSON(orderedVideos)
}
