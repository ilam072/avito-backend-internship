package domain

import "errors"

var (
	ErrTeamExists           = errors.New("team exists")
	ErrTeamNotFound         = errors.New("team not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrPullRequestExists    = errors.New("pull requests exists")
	ErrPullRequestNotFound  = errors.New("pull request not found")
	ErrPullRequestMerged    = errors.New("pull request merged")
	ErrUserNotAssignedForPR = errors.New("user isn't assigned for pr")
	ErrNoCandidate          = errors.New("no active candidate for pr")
)
