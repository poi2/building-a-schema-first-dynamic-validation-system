package model

import "time"

// User represents a user in the system
type User struct {
	ID        string
	Name      string
	Email     string
	Plan      string // free, pro, enterprise (stored as string in DB)
	CreatedAt time.Time
	UpdatedAt time.Time
}
