package models

import "time"

type Comment struct {
	ID        string    `json:"id"         db:"id"`
	VideoID   string    `json:"video_id"   db:"video_id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	Content   string    `json:"content"    db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
