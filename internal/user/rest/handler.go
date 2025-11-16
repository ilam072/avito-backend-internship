package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilam072/avito-backend-internship/internal/response"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/internal/types/dto"
	"github.com/rs/zerolog/log"
	"net/http"
)

type User interface {
	SetIsActive(ctx context.Context, ID uuid.UUID, isActive bool) (dto.UpdateUserResponse, error)
	GetUsersByTeam(ctx context.Context, name string) (dto.TeamWithMembers, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type UserHandler struct {
	user      User
	validator Validator
}

func NewUserHandler(user User, validator Validator) *UserHandler {
	return &UserHandler{user: user, validator: validator}
}

func (h *UserHandler) SetUserIsActive(c *gin.Context) {
	var req dto.SetUserIsActiveRequest
	if err := c.BindJSON(&req); err != nil {
		log.Logger.Warn().Err(err).Msg("failed to bind json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		log.Logger.Warn().Err(err).Msg("validation error")
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	user, err := h.user.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.NotFound(c)
			return
		}
		log.Logger.Error().Err(err).Any("req", req).Msg("failed to set user's is_active")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetTeam(c *gin.Context) {
	name := c.Query("team_name")
	if name == "" {
		response.BadRequest(c, "missing query param 'team_name'")
		return
	}

	team, err := h.user.GetUsersByTeam(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			response.NotFound(c)
			return
		}
		log.Logger.Error().Err(err).Any("name", name).Msg("failed to get team")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, team)
}
