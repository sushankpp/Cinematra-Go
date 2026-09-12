package models

import "time"

type Video struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id" db:"user_id"`
	UserName             string    `json:"user_name"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	ThumbnailUrl         string    `json:"thumbnail_url" db:"thumbnail_url"`
	StorageUrl           string    `json:"storage_url" db:"storage_url"`
	Status               string    `json:"status"`
	ModerationLabel      string    `json:"moderation_label" db:"moderation_label"`
	ModerationConfidence float64   `json:"moderation_confidence" db:"moderation_confidence"`
	CreatedAt            time.Time `json:"created_at"`
}
