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

type Team interface {
	CreateTeam(ctx context.Context, team dto.TeamWithMembers) (dto.TeamWithMembers, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type TeamHandler struct {
	team      Team
	validator Validator
}

func NewTeamHandler(team Team, validator Validator) *TeamHandler {
	return &TeamHandler{team: team, validator: validator}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var team dto.TeamWithMembers
	if err := c.BindJSON(&team); err != nil {
		log.Logger.Warn().Err(err).Msg("failed to bind team json")
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validator.Validate(team); err != nil {
		log.Logger.Warn().Err(err).Msg("validation error")
		response.BadRequest(c, fmt.Sprintf("validation error: %s", err.Error()))
		return
	}

	teamResp, err := h.team.CreateTeam(c.Request.Context(), team)
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			response.Conflict(c, "TEAM_EXISTS", "team_name already exists")
			return
		}
		log.Logger.Error().Err(err).Any("team", team).Msg("failed to create team")
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": teamResp})
}
