package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	prrepo "github.com/ilam072/avito-backend-internship/internal/pullrequest/repo"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/internal/types/dto"
	userrepo "github.com/ilam072/avito-backend-internship/internal/user/repo"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
)

type PullRequestRepo interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	GetPullRequestByID(ctx context.Context, ID uuid.UUID) (domain.PullRequest, error)
	GetPullRequestReviewers(ctx context.Context, ID uuid.UUID) ([]uuid.UUID, error)
	GetPRsWhereUserIsReviewer(ctx context.Context, userID uuid.UUID) ([]domain.PullRequest, error)
	UpdateReviewer(ctx context.Context, prID, oldUserID, newUserID uuid.UUID) error
	MergePullRequest(ctx context.Context, ID uuid.UUID) (domain.PullRequest, error)
	IsUserAssignedForPR(ctx context.Context, prID, userID uuid.UUID) (bool, error)
	PullRequestExists(ctx context.Context, id uuid.UUID) (bool, error)
}

type UserRepo interface {
	GetUserByID(ctx context.Context, ID uuid.UUID) (domain.User, error)
	GetNewUserIDForPRReview(ctx context.Context, oldUserID uuid.UUID) (uuid.UUID, error)
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
}

type PullRequest struct {
	userRepo UserRepo
	prRepo   PullRequestRepo
}

func NewPullRequest(userRepo UserRepo, prRepo PullRequestRepo) *PullRequest {
	return &PullRequest{userRepo: userRepo, prRepo: prRepo}
}

func (p *PullRequest) CreatePullRequest(ctx context.Context, pr dto.CreatePullRequest) (dto.GetPullRequest, error) {
	const op = "service.pr.Create"

	exists, err := p.prRepo.PullRequestExists(ctx, pr.ID)
	if err != nil {
		return dto.GetPullRequest{}, errutils.Wrap(op, err)
	}
	if exists {
		return dto.GetPullRequest{}, errutils.Wrap(op, domain.ErrPullRequestExists)
	}

	exists, err = p.userRepo.UserExists(ctx, pr.AuthorID)
	if err != nil {
		return dto.GetPullRequest{}, errutils.Wrap(op, err)
	}
	if !exists {
		return dto.GetPullRequest{}, errutils.Wrap(op, domain.ErrUserNotFound)
	}

	prDomain, err := p.prRepo.CreatePullRequest(ctx, domain.PullRequest{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
	})

	return dto.GetPullRequest{
		ID:        prDomain.ID,
		Name:      prDomain.Name,
		AuthorID:  prDomain.AuthorID,
		Status:    prDomain.Status,
		Reviewers: prDomain.Reviewers,
	}, nil
}

func (p *PullRequest) MergePullRequest(ctx context.Context, ID uuid.UUID) (dto.PRResponse, error) {
	const op = "service.pr.Merge"

	pr, err := p.prRepo.MergePullRequest(ctx, ID)
	if err != nil {
		if errors.Is(err, prrepo.ErrPullRequestNotFound) {
			return dto.PRResponse{}, errutils.Wrap(op, domain.ErrPullRequestNotFound)
		}
		return dto.PRResponse{}, errutils.Wrap(op, err)
	}

	return dto.PRResponse{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    pr.Status,
		Reviewers: pr.Reviewers,
		MergedAt:  *pr.MergedAt,
	}, nil
}

func (p *PullRequest) ReassignReviewer(ctx context.Context, prID uuid.UUID, userID uuid.UUID) (dto.ReassignResponse, error) {
	const op = "service.pr.ReassignReviewer"

	// prRepo.GetPullRequestByID (проверка merged, пр не найден)
	pr, err := p.prRepo.GetPullRequestByID(ctx, prID)
	if err != nil {
		if errors.Is(err, prrepo.ErrPullRequestNotFound) {
			return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrPullRequestNotFound)
		}
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}
	if pr.Status == "MERGED" {
		return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrPullRequestMerged)
	}

	// userRepo.GetUserByID (проверка пользователь не найден)
	user, err := p.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrUserNotFound)
		}
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}

	// prRepo.IsUserAssignedForPR (проверка что юзер назначен на пр)
	assigned, err := p.prRepo.IsUserAssignedForPR(ctx, prID, user.ID)
	if err != nil {
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}
	if !assigned {
		return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrUserNotAssignedForPR)
	}

	// userRepo.GetNewUserIDForPRReview (нет кандидатов)
	newUserID, err := p.userRepo.GetNewUserIDForPRReview(ctx, userID)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrNoCandidate)
		}
		return dto.ReassignResponse{}, err
	}

	// prRepo.UpdateReviewer
	if err := p.prRepo.UpdateReviewer(ctx, prID, userID, newUserID); err != nil {
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}

	// prRepo.GetPR
	pr, err = p.prRepo.GetPullRequestByID(ctx, prID)
	if err != nil {
		if errors.Is(err, prrepo.ErrPullRequestNotFound) {
			return dto.ReassignResponse{}, errutils.Wrap(op, domain.ErrPullRequestNotFound)
		}
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}

	// prRepo.GetReviewers
	pr.Reviewers, err = p.prRepo.GetPullRequestReviewers(ctx, prID)
	if err != nil {
		return dto.ReassignResponse{}, errutils.Wrap(op, err)
	}

	return dto.ReassignResponse{
		PR: dto.GetPullRequest{
			ID:        pr.ID,
			Name:      pr.Name,
			AuthorID:  pr.AuthorID,
			Status:    pr.Status,
			Reviewers: pr.Reviewers,
		},
		ReplacedBy: newUserID,
	}, nil
}

func (p *PullRequest) GetPRsWhereUserIsReviewer(ctx context.Context, userID uuid.UUID) (dto.GetReviewResponse, error) {
	const op = "service.pr.GetPRsWhereUserIsReviewer"

	prs, err := p.prRepo.GetPRsWhereUserIsReviewer(ctx, userID)
	if err != nil {
		return dto.GetReviewResponse{}, errutils.Wrap(op, err)
	}

	pullRequestsToReturn := make([]struct {
		ID       uuid.UUID `json:"pull_request_id"`
		Name     string    `json:"pull_request_name"`
		AuthorID uuid.UUID `json:"author_id"`
		Status   string    `json:"status"`
	}, len(prs))

	for i, pr := range prs {
		pullRequestsToReturn[i] = struct {
			ID       uuid.UUID `json:"pull_request_id"`
			Name     string    `json:"pull_request_name"`
			AuthorID uuid.UUID `json:"author_id"`
			Status   string    `json:"status"`
		}{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		}
	}

	response := dto.GetReviewResponse{
		UserID:       userID,
		PullRequests: pullRequestsToReturn,
	}

	return response, nil
}
