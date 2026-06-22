package recommend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

//this file is used to fetch every video with its genres and tags and then built a combined document for each video. This combined document will be then used by tokenize to process.

type rawVideoRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoGenres []struct {
		Genres struct {
			Name string `json:"name"`
		} `json:"genres"`
	} `json:"video_genres"`
	VideoTags []struct {
		Tags struct {
			Name string `json:"name"`
		} `json:"tags"`
	} `json:"video_tags"`
}

func FetchVideoDocs() ([]VideoDoc, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	url := supabaseURL + "/rest/v1/videos?select=id,title,description,video_genres(genres(name)),video_tags(tags(name))"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rows []rawVideoRow
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("failed to parse video rows: %w", err)
	}

	var docs []VideoDoc
	for _, row := range rows {
		var doc strings.Builder

		doc.WriteString(row.Title)
		doc.WriteString(" ")
		doc.WriteString(row.Title)
		doc.WriteString(" ")

		for _, genre := range row.VideoGenres {
			name := genre.Genres.Name
			doc.WriteString(name)
			doc.WriteString(" ")
			doc.WriteString(name)
			doc.WriteString(" ")
			doc.WriteString(name)
			doc.WriteString(" ")
		}

		for _, tag := range row.VideoTags {
			name := tag.Tags.Name
			doc.WriteString(name)
			doc.WriteString(" ")
			doc.WriteString(name)
			doc.WriteString(" ")
		}

		doc.WriteString(row.Description)

		docs = append(docs, VideoDoc{
			VideoId:      row.ID,
			DocumentText: doc.String(),
		})
	}

	return docs, nil
}
