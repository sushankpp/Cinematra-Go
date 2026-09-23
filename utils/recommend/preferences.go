package recommend

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type UserPreference struct {
	PreferredGenres []int `json:"preferred_genres"`
	PreferredTags   []int `json:"preferred_tags"`
}

var (
	genreCache   map[int]string
	tagCache     map[int]string
	cacheMutex   sync.RWMutex
	lastCacheHit time.Time
)

// FetchUserPreferences retrieves a user's onboarding preferences from Supabase
func FetchUserPreferences(userID string) (*UserPreference, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := fmt.Sprintf("%s/rest/v1/user_preferences?user_id=eq.%s&select=preferred_genres,preferred_tags", supabaseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Accept", "application/vnd.pgrst.object+json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == 406 {
		return nil, nil // No preferences set yet
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user preferences: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pref UserPreference
	if err := json.Unmarshal(body, &pref); err != nil {
		return nil, err
	}

	return &pref, nil
}

// FetchGenreMap returns a map of genre ID -> Name, cached in memory.
func FetchGenreMap() (map[int]string, error) {
	cacheMutex.RLock()
	if genreCache != nil && time.Since(lastCacheHit) < 1*time.Hour {
		defer cacheMutex.RUnlock()
		return genreCache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if genreCache != nil && time.Since(lastCacheHit) < 1*time.Hour {
		return genreCache, nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := supabaseURL + "/rest/v1/genres?select=id,name"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}

	newMap := make(map[int]string, len(rows))
	for _, r := range rows {
		newMap[r.ID] = r.Name
	}

	genreCache = newMap
	lastCacheHit = time.Now()
	return genreCache, nil
}

// FetchTagMap returns a map of tag ID -> Name, cached in memory.
func FetchTagMap() (map[int]string, error) {
	cacheMutex.RLock()
	if tagCache != nil && time.Since(lastCacheHit) < 1*time.Hour {
		defer cacheMutex.RUnlock()
		return tagCache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if tagCache != nil && time.Since(lastCacheHit) < 1*time.Hour {
		return tagCache, nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := supabaseURL + "/rest/v1/tags?select=id,name"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}

	newMap := make(map[int]string, len(rows))
	for _, r := range rows {
		newMap[r.ID] = r.Name
	}

	tagCache = newMap
	lastCacheHit = time.Now()
	return tagCache, nil
}

// BuildSeedProfile constructs a normalized TF-IDF taste profile using preferred genre and tag names.
// Genres are boosted 3x and tags are boosted 2x, consistent with corpus document weighting.
func BuildSeedProfile(genreNames, tagNames []string, idf map[string]float64) map[string]float64 {
	var doc strings.Builder

	for _, g := range genreNames {
		doc.WriteString(g)
		doc.WriteString(" ")
		doc.WriteString(g)
		doc.WriteString(" ")
		doc.WriteString(g)
		doc.WriteString(" ")
	}

	for _, t := range tagNames {
		doc.WriteString(t)
		doc.WriteString(" ")
		doc.WriteString(t)
		doc.WriteString(" ")
	}

	tokens := Tokenize(doc.String())
	if len(tokens) == 0 {
		return nil
	}

	vector := VectorizeTokens(tokens, idf)
	if len(vector) == 0 {
		return nil
	}

	// L2-normalize seed vector
	var magnitude float64
	for _, score := range vector {
		magnitude += score * score
	}
	magnitude = math.Sqrt(magnitude)
	if magnitude > 0 {
		for k := range vector {
			vector[k] /= magnitude
		}
	}

	return vector
}
