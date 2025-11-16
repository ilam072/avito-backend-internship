package repo

import (
	"context"
	"errors"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

var (
	ErrTeamExists   = errors.New("team exists")
	ErrTeamNotFound = errors.New("team not found")
)

func (r *TeamRepo) CreateTeam(ctx context.Context, name string, users []domain.User) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errutils.Wrap("failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO teams (name)
        VALUES ($1)
        RETURNING id;
	`
	var teamID int
	if err := tx.QueryRow(ctx, query, name).Scan(&teamID); err != nil {
		if isUniqueViolation(err) {
			return ErrTeamExists
		}
		return errutils.Wrap("failed to create team", err)
	}

	query = `
        INSERT INTO users (id, name, is_active, team_id)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) 
        DO UPDATE SET
            name = EXCLUDED.name,
            is_active = EXCLUDED.is_active,
            team_id = EXCLUDED.team_id; 
    `
	for _, u := range users {
		if _, err := tx.Exec(ctx, query, u.ID, u.Username, u.IsActive, teamID); err != nil {
			return errutils.Wrap("failed to create user", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return errutils.Wrap("failed to commit transaction", err)
	}

	return nil
}

func (r *TeamRepo) GetTeamNameByID(ctx context.Context, ID int) (string, error) {
	query := `
		SELECT name
		FROM teams
		WHERE id = $1;
	`

	var name string
	if err := r.db.QueryRow(ctx, query, ID).Scan(&name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errutils.Wrap("failed to get team name", ErrTeamNotFound)
		}
		return "", errutils.Wrap("failed to get team name", err)
	}
	return name, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
