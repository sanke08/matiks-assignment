package models

import "time"

type User struct {
	ID       int
	Username string
	Rating   int

	CreatedAt time.Time
	UpdatedAt time.Time
}
