package models

import "time"

type User struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Phone      *string   `json:"phone"`
	Password   *string   `json:"-"`
	Picture    *string   `json:"picture"`
	Provider   string    `json:"provider"`
	ProviderID *string   `json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
