package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilam072/avito-backend-internship/internal/response"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/internal/types/dto"
	"github.com/rs/zerolog/log"
	"net/http"
)

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pr dto.CreatePullRequest) (dto.GetPullRequest, error)
	MergePullRequest(ctx context.Context, ID string) (dto.PRResponse, error)
	ReassignReviewer(ctx context.Context, prID string, userID string) (dto.ReassignResponse, error)
	GetPRsWhereUserIsReviewer(ctx context.Context, userID string) (dto.GetReviewResponse, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type PullRequestHandler struct {
	pr        PullRequest
	validator Validator
}

func NewPullRequestHandler(pr PullRequest, validator Validator) *PullRequestHandler {
	return &PullRequestHandler{pr: pr, validator: validator}
}

func (h *PullRequestHandler) CreatePullRequest(c *gin.Context) {
	var pr dto.CreatePullRequest
	if err := c.BindJSON(&pr); err != nil {
		log.Logger.Warn().Err(err).Msg("failed to bind json to create pr req")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(pr); err != nil {
		log.Logger.Warn().Err(err).Msg("validation error")
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	prResp, err := h.pr.CreatePullRequest(c.Request.Context(), pr)
	if err != nil {
		if errors.Is(err, domain.ErrPullRequestExists) {
			response.Conflict(c, "PR_EXISTS", "PR id already exists")
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			response.NotFound(c)
			return
		}
		log.Logger.Error().Err(err).Any("pr", pr).Msg("failed to create pull request")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": prResp})
}

func (h *PullRequestHandler) MergePullRequest(c *gin.Context) {
	var req dto.MergePRRequest
	if err := c.BindJSON(&req); err != nil {
		log.Logger.Warn().Err(err).Msg("failed to bind json to merge pr req")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		log.Logger.Warn().Err(err).Msg("validation error")
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	prResp, err := h.pr.MergePullRequest(c.Request.Context(), req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrPullRequestNotFound) {
			response.NotFound(c)
			return
		}
		log.Logger.Error().Err(err).Any("req", req).Msg("failed to merge pull request")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": prResp})
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req dto.ReassignRequest
	if err := c.BindJSON(&req); err != nil {
		log.Logger.Warn().Err(err).Msg("failed to bind json to reassign req")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		log.Logger.Warn().Err(err).Msg("validation error")
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	prResp, err := h.pr.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrPullRequestNotFound) {
			response.NotFound(c)
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			response.NotFound(c)
			return
		}
		if errors.Is(err, domain.ErrPullRequestMerged) {
			response.Conflict(c, "PR_MERGED", "cannot reassign on merged PR")
			return
		}
		if errors.Is(err, domain.ErrUserNotAssignedForPR) {
			response.Conflict(c, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
			return
		}
		if errors.Is(err, domain.ErrNoCandidate) {
			response.Conflict(c, "NO_CANDIDATE", "no active replacement candidate in team")
			return
		}
		log.Logger.Error().Err(err).Any("req", req).Msg("failed to reassign reviewer")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, prResp)
}

func (h *PullRequestHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		log.Logger.Warn().Msg("empty user id")
		response.BadRequest(c, "invalid 'user_id' query parameter")
	}

	prsResp, err := h.pr.GetPRsWhereUserIsReviewer(c.Request.Context(), userID)
	if err != nil {
		log.Logger.Error().Err(err).Any("user_id", userID).Msg("failed to get prs by user id")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, prsResp)
}
