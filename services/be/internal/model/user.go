package model

import "time"

// User represents a user in the system
type User struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	Email     string    `yaml:"email"`
	Plan      string    `yaml:"plan"` // free, pro, enterprise
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}
