package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreatePullRequest struct {
	ID       uuid.UUID `json:"pull_request_id" validate:"required,uuid"`
	Name     string    `json:"pull_request_name" validate:"required"`
	AuthorID uuid.UUID `json:"author_id" validate:"required"`
}

type GetPullRequest struct {
	ID        uuid.UUID   `json:"pull_request_id"`
	Name      string      `json:"pull_request_name"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Status    string      `json:"status"`
	Reviewers []uuid.UUID `json:"assigned_reviewers"`
}

type MergePRRequest struct {
	ID uuid.UUID `json:"pull_request_id" validate:"required,uuid"`
}

type PRResponse struct {
	ID        uuid.UUID   `json:"pull_request_id"`
	Name      string      `json:"pull_request_name"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Status    string      `json:"status"`
	Reviewers []uuid.UUID `json:"assigned_reviewers"`
	MergedAt  time.Time   `json:"merged_at"`
}

type ReassignRequest struct {
	PullRequestID uuid.UUID `json:"pull_request_id" validate:"required,uuid"`
	UserID        uuid.UUID `json:"old_user_id" validate:"required,uuid"`
}

type ReassignResponse struct {
	PR         GetPullRequest `json:"pr"`
	ReplacedBy uuid.UUID      `json:"replaced_by"`
}

type GetReviewResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	PullRequests []struct {
		ID       uuid.UUID `json:"pull_request_id"`
		Name     string    `json:"pull_request_name"`
		AuthorID uuid.UUID `json:"author_id"`
		Status   string    `json:"status"`
	} `json:"pull_requests"`
}
