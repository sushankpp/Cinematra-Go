package recommend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
