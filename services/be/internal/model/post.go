package model

import "time"

// Post represents a post in the system
type Post struct {
	ID        string    `yaml:"id"`
	UserID    string    `yaml:"user_id"`
	Title     string    `yaml:"title"`
	Content   string    `yaml:"content"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}
