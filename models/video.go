package models

import "time"

type Video struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailUrl string    `json:"thumbnail_url" db:"thumbnail_url"`
	StorageUrl   string    `json:"storage_url" db:"storage_url"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
