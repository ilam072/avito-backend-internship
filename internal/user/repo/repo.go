package repo

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (r *UserRepo) GetUsersByTeam(ctx context.Context, name string) ([]domain.User, error) {
	query := `
		SELECT u.id, u.name, u.is_active, u.team_id, u.created_at, u.updated_at
		FROM users u
		JOIN teams t ON t.id = u.team_id
		WHERE t.name = $1;
	`

	rows, err := r.db.Query(ctx, query, name)
	if err != nil {
		return nil, errutils.Wrap("failed to query users by team", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.TeamID, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, errutils.Wrap("failed to scan user", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, errutils.Wrap("rows iteration error", err)
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	return users, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, ID uuid.UUID) (domain.User, error) {
	query := `
		SELECT id, name, is_active, team_id, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var user domain.User
	if err := r.db.QueryRow(ctx, query, ID).Scan(
		&user.ID,
		&user.Username,
		&user.IsActive,
		&user.TeamID,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, errutils.Wrap("failed to get user", err)
	}

	return user, nil
}

func (r *UserRepo) GetNewUserIDForPRReview(ctx context.Context, prID uuid.UUID, oldUserID uuid.UUID) (uuid.UUID, error) {
	query := `
   		SELECT id 
   		FROM users u 
   		WHERE u.is_active = TRUE 
   		  AND u.team_id = (SELECT team_id FROM users u2 WHERE u2.id = $1) 
   		  AND u.id NOT IN (
   		      SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $2
   		  )
   		ORDER BY random() 
   		LIMIT 1
	`
	var ID uuid.UUID
	if err := r.db.QueryRow(ctx, query, oldUserID, prID).Scan(&ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrUserNotFound
		}
		return uuid.Nil, errutils.Wrap("failed to get new user id", err)
	}

	return ID, nil
}
func (r *UserRepo) UpdateIsActive(ctx context.Context, ID uuid.UUID, isActive bool) error {
	query := `
		UPDATE users
		SET is_active = $1,
		    updated_at = NOW()
		WHERE id = $2;
	`

	res, err := r.db.Exec(ctx, query, isActive, ID)
	if err != nil {
		return errutils.Wrap("failed to update user is_active", err)
	}

	if rows := res.RowsAffected(); rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) UserExists(ctx context.Context, ID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`
	var exists bool

	if err := r.db.QueryRow(ctx, query, ID).Scan(&exists); err != nil {
		return false, errutils.Wrap("failed to check existing user", err)
	}

	return exists, nil
}
