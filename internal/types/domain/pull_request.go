package domain

import (
	"time"
)

type PullRequest struct {
	ID        string
	AuthorID  string
	Name      string
	Status    string
	Reviewers []string
	CreatedAt time.Time
	MergedAt  *time.Time
}
