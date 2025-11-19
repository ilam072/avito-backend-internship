package dto

import (
	"time"
)

type CreatePullRequest struct {
	ID       string `json:"pull_request_id" validate:"required"`
	Name     string `json:"pull_request_name" validate:"required"`
	AuthorID string `json:"author_id" validate:"required"`
}

type GetPullRequest struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    string   `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

type MergePRRequest struct {
	ID string `json:"pull_request_id" validate:"required"`
}

type PRResponse struct {
	ID        string    `json:"pull_request_id"`
	Name      string    `json:"pull_request_name"`
	AuthorID  string    `json:"author_id"`
	Status    string    `json:"status"`
	Reviewers []string  `json:"assigned_reviewers"`
	MergedAt  time.Time `json:"merged_at"`
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	UserID        string `json:"old_user_id" validate:"required"`
}

type ReassignResponse struct {
	PR         GetPullRequest `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type GetReviewResponse struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		ID       string `json:"pull_request_id"`
		Name     string `json:"pull_request_name"`
		AuthorID string `json:"author_id"`
		Status   string `json:"status"`
	} `json:"pull_requests"`
}
