package domain

import (
	"time"
)

type User struct {
	ID        string
	TeamID    int
	Username  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
