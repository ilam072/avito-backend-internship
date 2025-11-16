package service

import (
	"context"
	"errors"
	"github.com/ilam072/avito-backend-internship/internal/team/repo"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/internal/types/dto"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
)

type TeamRepo interface {
	CreateTeam(ctx context.Context, name string, users []domain.User) error
	//GetTeam(ctx context.Context, name string) ([]domain.User, error)
}

type Team struct {
	repo TeamRepo
}

func NewTeam(repo TeamRepo) *Team {
	return &Team{repo: repo}
}

func (t *Team) CreateTeam(ctx context.Context, team dto.TeamWithMembers) (dto.TeamWithMembers, error) {
	const op = "service.team.Create"

	users := make([]domain.User, len(team.Members))
	for i, member := range team.Members {
		users[i] = domain.User{
			ID:       member.ID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	if err := t.repo.CreateTeam(ctx, team.TeamName, users); err != nil {
		if errors.Is(err, repo.ErrTeamExists) {
			return dto.TeamWithMembers{}, errutils.Wrap(op, domain.ErrTeamExists)
		}
		return dto.TeamWithMembers{}, errutils.Wrap(op, err)
	}

	return team, nil
}

/*func (t *Team) GetTeam(ctx context.Context, name string) (dto.TeamWithMembers, error) {
	const op = "service.team.Create"

	users, err := t.repo.GetTeam(ctx, name)
	if err != nil {
		if errors.Is(err, repo.ErrTeamNotFound) {
			return dto.TeamWithMembers{}, errutils.Wrap(op, domain.ErrTeamNotFound)
		}
		return dto.TeamWithMembers{}, errutils.Wrap(op, err)
	}

	members := make([]dto.User, len(users))
	for i, user := range users {
		members[i] = dto.User{
			ID:       user.ID,
			Username: user.Username,
			IsActive: user.IsActive,
		}
	}

	return dto.TeamWithMembers{
		TeamName: name,
		Members:  members,
	}, nil
}
*/
