package domain

import (
	"github.com/google/uuid"
	"time"
)

type PullRequest struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Name      string
	Status    string
	Reviewers []uuid.UUID
	CreatedAt time.Time
	MergedAt  *time.Time
}
