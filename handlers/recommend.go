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

	"github.com/gofiber/fiber/v3"
)

// GetRecommendations handles recommendations for the current user
func GetRecommendations(c fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Fetch all videos from the database to build the corpus
	docs, err := recommend.FetchVideoDocs()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch videos from database"})
	}

	corpus := recommend.BuildCorpus(docs)
	if corpus == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to build video corpus"})
	}

	// Fetch user's watch history
	watched, err := recommend.FetchWatchHistory(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch watch history"})
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
			return nil, fmt.Errorf("database returned status %s: %s", resp.Status, string(body))
		}
		return body, nil
	}

	// Cold start: User has no watch history — return latest videos with similarity_score: 0
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
			coldStartVideos[i]["similarity_score"] = 0.0
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

		// Attach similarity scores to fallback videos
		for i, v := range fallbackVideos {
			if id, ok := v["id"].(string); ok {
				if vec, exists := corpus.VideoVectors[id]; exists {
					sim := recommend.CosineSimilarity(userProfile, vec.Vector)
					fallbackVideos[i]["similarity_score"] = math.Round(sim*10000) / 10000
				} else {
					fallbackVideos[i]["similarity_score"] = 0.0
				}
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
	// and attach similarity_score to each video
	videoMap := make(map[string]map[string]interface{})
	for _, v := range videos {
		if id, ok := v["id"].(string); ok {
			videoMap[id] = v
		}
	}

	orderedVideos := make([]map[string]interface{}, 0, len(recommendedIDs))
	for _, id := range recommendedIDs {
		if v, exists := videoMap[id]; exists {
			v["similarity_score"] = math.Round(scoreMap[id]*10000) / 10000 // round to 4 decimal places
			orderedVideos = append(orderedVideos, v)
		}
	}

	return c.Status(200).JSON(orderedVideos)
}
