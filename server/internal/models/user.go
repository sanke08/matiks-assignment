package models

import "time"

type User struct {
	ID       string
	Username string
	Rating   int

	CreatedAt time.Time
	UpdatedAt time.Time
}
