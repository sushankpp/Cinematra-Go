package handlers

import (
	"cinematra-backend/utils/recommend"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// fetchVideosByIDs fetches full video details from Supabase for a given list of IDs
func fetchVideosByIDs(supabaseURL, serviceKey string, ids []string) ([]map[string]interface{}, error) {
	url := supabaseURL + "/rest/v1/videos" +
		"?id=in.(" + strings.Join(ids, ",") + ")" +
		"&status=eq.ready" +
		"&select=id,title,description,thumbnail_url,user_id,release_year,language,content_rating,moderation_label,moderation_confidence,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))"

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

	var videos []map[string]interface{}
	if err := json.Unmarshal(body, &videos); err != nil {
		return nil, err
	}

	// Reorder results to match the ranked IDs order
	videoMap := make(map[string]map[string]interface{})
	for _, v := range videos {
		if id, ok := v["id"].(string); ok {
			videoMap[id] = v
		}
	}

	ordered := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		if v, exists := videoMap[id]; exists {
			ordered = append(ordered, v)
		}
	}

	return ordered, nil
}

// fetchLatestVideos is the fallback — returns the latest ready videos ordered by created_at
func fetchLatestVideos(supabaseURL, serviceKey string, limit int) ([]byte, error) {
	url := supabaseURL + "/rest/v1/videos" +
		"?status=eq.ready" +
		"&select=id,title,thumbnail_url,user_id,release_year,created_at,views,users(first_name,last_name,picture),video_genres(genre_id,genres(id,name,slug)),video_tags(tag_id,tags(id,name,slug))" +
		"&order=created_at.desc" +
		"&limit=10"

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
	return body, err
}

// GetGlobalTrending returns the top trending videos globally based on the last 7 days of watch activity.
// Public endpoint — no auth required.
func GetGlobalTrending(c fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Fetch trending scores from the last 7 days
	scores, err := recommend.FetchTrendingScore()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch trending scores: " + err.Error()})
	}

	// No trending data yet → fall back to latest videos
	if len(scores) == 0 {
		body, err := fetchLatestVideos(supabaseURL, serviceKey, 10)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch latest videos: " + err.Error()})
		}
		return c.Status(200).Type("json").Send(body)
	}

	// Get top 10 ranked video IDs by trending score
	rankedIDs := recommend.RankTrending(scores, 10)

	// Fetch full video details for those IDs in ranked order
	videos, err := fetchVideosByIDs(supabaseURL, serviceKey, rankedIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch video details: " + err.Error()})
	}

	return c.Status(200).JSON(videos)
}

// GetPersonalizedTrending returns trending videos blended with the user's content preferences.
// Auth required.
func GetPersonalizedTrending(c fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Fetch user watch history
	watched, err := recommend.FetchWatchHistory(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch watch history: " + err.Error()})
	}

	corpus, trendingScores, maxTrending := recommend.Cache.Snapshot()
	if corpus == nil {
		return c.Status(503).JSON(fiber.Map{"error": "recommendations warming up, try again shortly"})
	}

	// Build user content preference profile (from watch history if available, or onboarding preferences)
	var userProfile map[string]float64
	if len(watched) > 0 {
		userProfile = recommend.BuildUserProfile(watched, corpus)
	} else {
		// Cold start: Check if user has onboarding preferences
		pref, _ := recommend.FetchUserPreferences(userID)
		if pref != nil && (len(pref.PreferredGenres) > 0 || len(pref.PreferredTags) > 0) {
			genreMap, _ := recommend.FetchGenreMap()
			tagMap, _ := recommend.FetchTagMap()

			var genreNames []string
			for _, gid := range pref.PreferredGenres {
				if name, ok := genreMap[gid]; ok {
					genreNames = append(genreNames, name)
				}
			}

			var tagNames []string
			for _, tid := range pref.PreferredTags {
				if name, ok := tagMap[tid]; ok {
					tagNames = append(tagNames, name)
				}
			}

			userProfile = recommend.BuildSeedProfile(genreNames, tagNames, corpus.IDF)
		}
	}

	// If user has neither watch history nor preferences, fall back to global trending
	if len(userProfile) == 0 {
		trendingScores, err := recommend.FetchTrendingScore()
		if err != nil || len(trendingScores) == 0 {
			return GetGlobalTrending(c)
		}

		maxT := 0.0
		for _, s := range trendingScores {
			if s > maxT {
				maxT = s
			}
		}

		rankedIDs := recommend.RankTrending(trendingScores, 10)
		videos, err := fetchVideosByIDs(supabaseURL, serviceKey, rankedIDs)
		if err != nil {
			return GetGlobalTrending(c)
		}

		for i, v := range videos {
			if id, ok := v["id"].(string); ok {
				ts := 0.0
				if s, exists := trendingScores[id]; exists && maxT > 0 {
					ts = s / maxT
				}
				videos[i]["similarity_score"] = 0.0
				videos[i]["trending_score"] = math.Round(ts*10000) / 10000
				videos[i]["combined_score"] = math.Round(0.3*ts*10000) / 10000
			}
		}

		return c.Status(200).JSON(videos)
	}

	// Build set of already-watched IDs to exclude from results
	watchedIDs := make(map[string]bool)
	for _, w := range watched {
		watchedIDs[w.VideoId] = true
	}

	// Score every unwatched video using hybrid formula
	type hybridEntry struct {
		videoId       string
		similarity    float64
		trendingScore float64
		combinedScore float64
	}

	var unwatchedCandidates []hybridEntry
	var watchedCandidates []hybridEntry
	for videoId, videoVector := range corpus.VideoVectors {
		similarity := recommend.CosineSimilarity(userProfile, videoVector.Vector)

		trendingScore := 0.0
		if s, exists := trendingScores[videoId]; exists && maxTrending > 0 {
			trendingScore = s / maxTrending
		}

		// 70% content similarity + 30% trending popularity
		combined := 0.7*similarity + 0.3*trendingScore

		if combined > 0 {
			entry := hybridEntry{
				videoId:       videoId,
				similarity:    similarity,
				trendingScore: trendingScore,
				combinedScore: combined,
			}
			if watchedIDs[videoId] {
				watchedCandidates = append(watchedCandidates, entry)
			} else {
				unwatchedCandidates = append(unwatchedCandidates, entry)
			}
		}
	}

	// Sort unwatched by combined score descending
	sort.Slice(unwatchedCandidates, func(i, j int) bool {
		return unwatchedCandidates[i].combinedScore > unwatchedCandidates[j].combinedScore
	})

	candidates := unwatchedCandidates
	if len(candidates) > 10 {
		candidates = candidates[:10]
	}

	// Backfill with top-scoring watched videos if unwatched catalog is exhausted
	if len(candidates) < 10 && len(watchedCandidates) > 0 {
		sort.Slice(watchedCandidates, func(i, j int) bool {
			return watchedCandidates[i].combinedScore > watchedCandidates[j].combinedScore
		})
		needed := 10 - len(candidates)
		if needed > len(watchedCandidates) {
			needed = len(watchedCandidates)
		}
		candidates = append(candidates, watchedCandidates[:needed]...)
	}

	// If no hybrid results, fall back to global trending but still compute scores
	if len(candidates) == 0 {
		trendingIDs := recommend.RankTrending(trendingScores, 10)

		// If no trending data at all, fetch latest
		if len(trendingIDs) == 0 {
			return GetGlobalTrending(c)
		}

		videos, err := fetchVideosByIDs(supabaseURL, serviceKey, trendingIDs)
		if err != nil {
			return GetGlobalTrending(c)
		}

		for i, v := range videos {
			if id, ok := v["id"].(string); ok {
				sim := 0.0
				if vec, exists := corpus.VideoVectors[id]; exists {
					sim = recommend.CosineSimilarity(userProfile, vec.Vector)
				}
				ts := 0.0
				if s, exists := trendingScores[id]; exists && maxTrending > 0 {
					ts = s / maxTrending
				}
				videos[i]["similarity_score"] = math.Round(sim*10000) / 10000
				videos[i]["trending_score"] = math.Round(ts*10000) / 10000
				videos[i]["combined_score"] = math.Round((0.7*sim+0.3*ts)*10000) / 10000
			}
		}

		return c.Status(200).JSON(videos)
	}

	// Take top 10 IDs
	limit := 10
	if len(candidates) < limit {
		limit = len(candidates)
	}
	rankedIDs := make([]string, 0, limit)
	scoreMap := make(map[string]hybridEntry, limit)
	for i := 0; i < limit; i++ {
		rankedIDs = append(rankedIDs, candidates[i].videoId)
		scoreMap[candidates[i].videoId] = candidates[i]
	}

	// Fetch full video details in ranked order
	videos, err := fetchVideosByIDs(supabaseURL, serviceKey, rankedIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch video details: " + err.Error()})
	}

	// Attach scores to each video
	for i, v := range videos {
		if id, ok := v["id"].(string); ok {
			if entry, exists := scoreMap[id]; exists {
				videos[i]["similarity_score"] = math.Round(entry.similarity*10000) / 10000
				videos[i]["trending_score"] = math.Round(entry.trendingScore*10000) / 10000
				videos[i]["combined_score"] = math.Round(entry.combinedScore*10000) / 10000
			}
		}
	}

	return c.Status(200).JSON(videos)
}
