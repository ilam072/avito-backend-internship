package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/internal/types/dto"
	"github.com/ilam072/avito-backend-internship/internal/user/repo"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
)

type TeamRepo interface {
	GetTeamNameByID(ctx context.Context, ID int) (string, error)
}

type UserRepo interface {
	GetUsersByTeam(ctx context.Context, name string) ([]domain.User, error)
	UpdateIsActive(ctx context.Context, ID uuid.UUID, isActive bool) error
	GetUserByID(ctx context.Context, ID uuid.UUID) (domain.User, error)
}

type User struct {
	userRepo UserRepo
	teamRepo TeamRepo
}

func NewUser(userRepo UserRepo, teamRepo TeamRepo) *User {
	return &User{userRepo: userRepo, teamRepo: teamRepo}
}

func (u *User) GetUsersByTeam(ctx context.Context, name string) (dto.TeamWithMembers, error) {
	const op = "service.user.GetUsersByTeam"

	users, err := u.userRepo.GetUsersByTeam(ctx, name)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
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

func (u *User) SetIsActive(ctx context.Context, ID uuid.UUID, isActive bool) (dto.UpdateUserResponse, error) {
	const op = "service.user.SetIsActive"

	if err := u.userRepo.UpdateIsActive(ctx, ID, isActive); err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return dto.UpdateUserResponse{}, errutils.Wrap(op, domain.ErrUserNotFound)
		}
		return dto.UpdateUserResponse{}, errutils.Wrap(op, err)
	}

	user, err := u.userRepo.GetUserByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return dto.UpdateUserResponse{}, errutils.Wrap(op, domain.ErrUserNotFound)
		}
		return dto.UpdateUserResponse{}, errutils.Wrap(op, err)
	}

	teamName, err := u.teamRepo.GetTeamNameByID(ctx, user.TeamID)
	if err != nil {
		return dto.UpdateUserResponse{}, errutils.Wrap(op, err)
	}

	return dto.UpdateUserResponse{
		ID:       user.ID,
		Username: user.Username,
		TeamName: teamName,
		IsActive: user.IsActive,
	}, nil
}
