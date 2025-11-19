package repo

import (
	"context"
	"errors"
	"github.com/ilam072/avito-backend-internship/internal/types/domain"
	"github.com/ilam072/avito-backend-internship/pkg/errutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestsRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *PullRequestsRepo {
	return &PullRequestsRepo{db: db}
}

var (
	ErrPullRequestNotFound = errors.New("pull request not found")
)

func (r *PullRequestsRepo) CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO pull_requests (id, author_id, name)
		VALUES ($1, $2, $3)
		RETURNING status, created_at
	`
	if err = tx.QueryRow(ctx, query, pr.ID, pr.AuthorID, pr.Name).Scan(&pr.Status, &pr.CreatedAt); err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to insert pull request", err)
	}

	query = `
		SELECT u.id
		FROM users u
		JOIN users a ON a.id = $1
		WHERE u.team_id = a.team_id
		  AND u.is_active = TRUE
		  AND u.id != $1
		ORDER BY RANDOM()
		LIMIT 2
	`

	rows, err := tx.Query(ctx, query, pr.AuthorID)
	if err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to select reviewers", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var rID string
		if err := rows.Scan(&rID); err != nil {
			return domain.PullRequest{}, errutils.Wrap("failed to scan reviewer", err)
		}
		reviewers = append(reviewers, rID)
	}

	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, errutils.Wrap("error iterating reviewers", err)
	}

	query = `
			INSERT INTO pr_reviewers (pr_id, reviewer_id)
			VALUES ($1, $2)
		`

	for _, rID := range reviewers {
		if _, err := tx.Exec(ctx, query, pr.ID, rID); err != nil {
			return domain.PullRequest{}, errutils.Wrap("failed to insert reviewer", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to commit transaction", err)
	}

	pr.Reviewers = make([]string, len(reviewers))
	for i, rID := range reviewers {
		pr.Reviewers[i] = rID
	}

	return pr, nil
}

func (r *PullRequestsRepo) GetPullRequestByID(ctx context.Context, ID string) (domain.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1;
	`

	var pr domain.PullRequest
	if err := r.db.QueryRow(ctx, query, ID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, ErrPullRequestNotFound
		}
		return domain.PullRequest{}, errutils.Wrap("failed to get pr", err)
	}

	return pr, nil

}

func (r *PullRequestsRepo) GetPullRequestReviewers(ctx context.Context, ID string) ([]string, error) {
	query := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pr_id = $1
	`

	rows, err := r.db.Query(ctx, query, ID)
	if err != nil {
		return nil, errutils.Wrap("failed to get reviewers", err)
	}
	defer rows.Close()

	var reviewersIDs []string
	for rows.Next() {
		var rID string
		if err := rows.Scan(&rID); err != nil {
			return nil, errutils.Wrap("failed to scan reviewer", err)
		}
		reviewersIDs = append(reviewersIDs, rID)
	}

	if err := rows.Err(); err != nil {
		return nil, errutils.Wrap("error iterating reviewers", err)
	}

	return reviewersIDs, nil
}

func (r *PullRequestsRepo) MergePullRequest(ctx context.Context, ID string) (domain.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr domain.PullRequest
	if err = tx.QueryRow(ctx, query, ID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, ErrPullRequestNotFound
		}
		return domain.PullRequest{}, errutils.Wrap("failed to get pull request", err)
	}

	if pr.Status != "MERGED" {
		query = `
			UPDATE pull_requests
			SET status = 'MERGED',
			    merged_at = NOW()
			WHERE id = $1
			RETURNING status, merged_at
		`
		if err = tx.QueryRow(ctx, query, ID).Scan(&pr.Status, &pr.MergedAt); err != nil {
			return domain.PullRequest{}, errutils.Wrap("failed to merge pull request", err)
		}
	}

	query = `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`
	rows, err := tx.Query(ctx, query, pr.ID)
	if err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to select reviewers", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rID string
		if err := rows.Scan(&rID); err != nil {
			return domain.PullRequest{}, errutils.Wrap("failed to scan reviewer", err)
		}
		pr.Reviewers = append(pr.Reviewers, rID)
	}
	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, errutils.Wrap("error iterating reviewers", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.PullRequest{}, errutils.Wrap("failed to commit transaction", err)
	}

	return pr, nil
}

func (r *PullRequestsRepo) IsUserAssignedForPR(ctx context.Context, prID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
    	SELECT 1
    	FROM pr_reviewers
    	WHERE pr_id = $1 AND reviewer_id = $2
		);
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, prID, userID).Scan(&exists); err != nil {
		return false, errutils.Wrap("failed to check if user exists", err)
	}

	return exists, nil
}

func (r *PullRequestsRepo) UpdateReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	query := `
		UPDATE pr_reviewers 
		SET reviewer_id = $1 
		WHERE pr_id = $2 AND reviewer_id = $3
	`

	if _, err := r.db.Exec(ctx, query, newUserID, prID, oldUserID); err != nil {
		return errutils.Wrap("failed to update reviewer", err)
	}

	return nil
}

func (r *PullRequestsRepo) GetPRsWhereUserIsReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	query := `
		SELECT 
			pr.id,
			pr.name,
			pr.author_id,
			pr.status,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		JOIN pr_reviewers r ON pr.id = r.pr_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errutils.Wrap("failed to get PRs where user is reviewer", err)
	}
	defer rows.Close()

	var prs []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest

		if err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		); err != nil {
			return nil, errutils.Wrap("failed to scan pr", err)
		}

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, errutils.Wrap("failed to iterate PR rows", err)
	}

	return prs, nil
}

func (r *PullRequestsRepo) PullRequestExists(ctx context.Context, ID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id=$1)`
	var exists bool

	if err := r.db.QueryRow(ctx, query, ID).Scan(&exists); err != nil {
		return false, errutils.Wrap("failed to check existing PR", err)
	}

	return exists, nil
}
