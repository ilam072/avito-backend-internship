package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	TeamID    int
	Username  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
