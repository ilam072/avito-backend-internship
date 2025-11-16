package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	prrepo "github.com/ilam072/avito-backend-internship/internal/pullrequest/repo"
	prrest "github.com/ilam072/avito-backend-internship/internal/pullrequest/rest"
	prservice "github.com/ilam072/avito-backend-internship/internal/pullrequest/service"
	"github.com/ilam072/avito-backend-internship/internal/router"
	teamrepo "github.com/ilam072/avito-backend-internship/internal/team/repo"
	teamrest "github.com/ilam072/avito-backend-internship/internal/team/rest"
	teamservice "github.com/ilam072/avito-backend-internship/internal/team/service"
	userrepo "github.com/ilam072/avito-backend-internship/internal/user/repo"
	userrest "github.com/ilam072/avito-backend-internship/internal/user/rest"
	userservice "github.com/ilam072/avito-backend-internship/internal/user/service"
	"github.com/ilam072/avito-backend-internship/internal/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testDBConnStr = "postgres://postgres:postgres@localhost:5433/pr_db_test?sslmode=disable"

func CleanDB(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", strings.Join(tables, ", "))
	_, err := db.Exec(ctx, query)
	require.NoError(t, err, "Failed to clean database")
}

func SetupRouterForTesting(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(testDBConnStr)
	if err != nil {
		t.Fatalf("Failed to parse DB config: %v", err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	CleanDB(t, dbPool)

	v := validator.New()

	userR := userrepo.New(dbPool)
	teamR := teamrepo.New(dbPool)
	prR := prrepo.New(dbPool)

	userS := userservice.NewUser(userR, teamR)
	teamS := teamservice.NewTeam(teamR)
	prS := prservice.NewPullRequest(userR, prR)

	userH := userrest.NewUserHandler(userS, v)
	teamH := teamrest.NewTeamHandler(teamS, v)
	prH := prrest.NewPullRequestHandler(prS, v)

	r := router.New(userH, teamH, prH)

	return r, dbPool
}

type TeamWithMembers struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		ID       uuid.UUID `json:"user_id"`
		Username string    `json:"username"`
		IsActive bool      `json:"is_active"`
	} `json:"members"`
}

func createTeamHTTP(t *testing.T, r *gin.Engine, team TeamWithMembers) {
	ctx := context.Background()
	body, _ := json.Marshal(team)
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, "POST", "/api/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "Precondition: failed to create team")
}

func createPRHTTP(t *testing.T, r *gin.Engine, prID uuid.UUID, name string, authorID uuid.UUID) GetPullRequest {
	ctx := context.Background()

	createReq := struct {
		ID       uuid.UUID `json:"pull_request_id"`
		Name     string    `json:"pull_request_name"`
		AuthorID uuid.UUID `json:"author_id"`
	}{ID: prID, Name: name, AuthorID: authorID}
	body, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "Precondition: failed to create PR")

	var resp struct {
		PR GetPullRequest `json:"pr"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp.PR
}

func addReviewerDirect(t *testing.T, ctx context.Context, db *pgxpool.Pool, prID, reviewerID uuid.UUID) {
	q := `INSERT INTO pr_reviewers (pr_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())`
	_, err := db.Exec(ctx, q, prID, reviewerID)
	if err != nil {
		t.Fatalf("failed to add reviewer direct: %v", err)
	}
}

type GetPullRequest struct {
	ID        uuid.UUID   `json:"pull_request_id"`
	Name      string      `json:"pull_request_name"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Status    string      `json:"status"`
	Reviewers []uuid.UUID `json:"assigned_reviewers"`
}

//
// === TESTS
//

func TestGetTeam(t *testing.T) {
	r, db := SetupRouterForTesting(t)
	ctx := context.Background()

	const teamName = "testing_squad_get"

	userID1 := uuid.New()
	userID2 := uuid.New()

	teamReq := TeamWithMembers{
		TeamName: teamName,
		Members: []struct {
			ID       uuid.UUID `json:"user_id"`
			Username string    `json:"username"`
			IsActive bool      `json:"is_active"`
		}{
			{ID: userID1, Username: "ActiveUser", IsActive: true},
			{ID: userID2, Username: "InactiveUser", IsActive: false},
		},
	}
	createTeamHTTP(t, r, teamReq)

	t.Run("GetTeam_Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "GET", "/api/team/get?team_name="+teamName, nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp TeamWithMembers
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, teamName, resp.TeamName)
		assert.Len(t, resp.Members, 2)

		var activeCount, inactiveCount int
		for _, member := range resp.Members {
			if member.IsActive {
				activeCount++
			} else {
				inactiveCount++
			}
		}
		assert.Equal(t, 1, activeCount, "Should be one active user")
		assert.Equal(t, 1, inactiveCount, "Should be one inactive user")
	})

	t.Run("GetTeam_NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "GET", "/api/team/get?team_name=non_existent", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		var errResp struct {
			Error struct {
				Code string `json:"code"`
			}
		}
		json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
	})

	db.Close()
}

func TestPullRequestCreate(t *testing.T) {
	r, db := SetupRouterForTesting(t)

	const teamName = "pr_squad"
	authorID := uuid.New()
	reviewerID1 := uuid.New()
	reviewerID2 := uuid.New()
	reviewerID3 := uuid.New()
	inactiveID := uuid.New()

	teamReq := TeamWithMembers{
		TeamName: teamName,
		Members: []struct {
			ID       uuid.UUID `json:"user_id"`
			Username string    `json:"username"`
			IsActive bool      `json:"is_active"`
		}{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: reviewerID1, Username: "Rev1", IsActive: true},
			{ID: reviewerID2, Username: "Rev2", IsActive: true},
			{ID: reviewerID3, Username: "Rev3", IsActive: true},
			{ID: inactiveID, Username: "Inactive", IsActive: false},
		},
	}
	createTeamHTTP(t, r, teamReq)

	t.Run("CreatePR_Success_AssignsTwoReviewers", func(t *testing.T) {
		prID := uuid.New()
		prName := "NewFeaturePR"

		pr := createPRHTTP(t, r, prID, prName, authorID)

		assert.Equal(t, prID, pr.ID)
		assert.Equal(t, authorID, pr.AuthorID)
		assert.Equal(t, "OPEN", pr.Status)
		// у нас логика сервиса назначает до 2, но т.к. random, мы просто проверим количество <=2 и автор/неактивный не в списке
		assert.LessOrEqual(t, len(pr.Reviewers), 2)
		for _, reviewerID := range pr.Reviewers {
			assert.NotEqual(t, authorID, reviewerID, "Автор не должен быть ревьюером")
			assert.NotEqual(t, inactiveID, reviewerID, "Неактивный пользователь не должен быть ревьюером")
		}
	})

	t.Run("CreatePR_Conflict_IDExists", func(t *testing.T) {
		prID := uuid.New()
		prName := "AnotherPR"

		_ = createPRHTTP(t, r, prID, prName, authorID)

		ctx := context.Background()
		createReq := struct {
			ID       uuid.UUID `json:"pull_request_id"`
			Name     string    `json:"pull_request_name"`
			AuthorID uuid.UUID `json:"author_id"`
		}{ID: prID, Name: prName, AuthorID: authorID}
		body, _ := json.Marshal(createReq)

		wConflict := httptest.NewRecorder()
		reqConflict, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/create", bytes.NewBuffer(body))
		reqConflict.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(wConflict, reqConflict)

		require.Equal(t, http.StatusConflict, wConflict.Code)
		var errResp struct {
			Error struct {
				Code string `json:"code"`
			}
		}
		json.Unmarshal(wConflict.Body.Bytes(), &errResp)
		assert.Equal(t, "PR_EXISTS", errResp.Error.Code)
	})

	db.Close()
}

func TestPullRequestMerge(t *testing.T) {
	r, db := SetupRouterForTesting(t)
	ctx := context.Background()

	const teamName = "merge_squad"
	authorID := uuid.New()
	prID := uuid.New()

	teamReq := TeamWithMembers{
		TeamName: teamName,
		Members: []struct {
			ID       uuid.UUID `json:"user_id"`
			Username string    `json:"username"`
			IsActive bool      `json:"is_active"`
		}{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: uuid.New(), Username: "Rev1", IsActive: true},
			{ID: uuid.New(), Username: "Rev2", IsActive: true},
		},
	}
	createTeamHTTP(t, r, teamReq)

	_ = createPRHTTP(t, r, prID, "MergeTarget", authorID)

	t.Run("MergePR_Success", func(t *testing.T) {
		mergeReq := struct {
			ID uuid.UUID `json:"pull_request_id"`
		}{ID: prID}
		body, _ := json.Marshal(mergeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/merge", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			PR struct {
				Status   string  `json:"status"`
				MergedAt *string `json:"merged_at"`
			} `json:"pr"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "MERGED", resp.PR.Status)
		assert.NotNil(t, resp.PR.MergedAt, "MergedAt должно быть установлено")
	})

	t.Run("MergePR_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		mergeReq := struct {
			ID uuid.UUID `json:"pull_request_id"`
		}{ID: nonExistentID}
		body, _ := json.Marshal(mergeReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/merge", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
	})

	db.Close()
}

func TestPullRequestReassign(t *testing.T) {
	r, db := SetupRouterForTesting(t)
	ctx := context.Background()

	const teamName = "reassign_squad"
	authorID := uuid.New()
	reviewerToReplace := uuid.New()
	newCandidateID := uuid.New()
	otherID := uuid.New()

	prID := uuid.New()

	teamReq := TeamWithMembers{
		TeamName: teamName,
		Members: []struct {
			ID       uuid.UUID `json:"user_id"`
			Username string    `json:"username"`
			IsActive bool      `json:"is_active"`
		}{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: reviewerToReplace, Username: "OldReviewer", IsActive: true},
			{ID: newCandidateID, Username: "NewCandidate", IsActive: true},
			{ID: otherID, Username: "RevOther", IsActive: true},
		},
	}
	createTeamHTTP(t, r, teamReq)

	_ = createPRHTTP(t, r, prID, "ReassignTarget", authorID)

	ctxDB := context.Background()
	_, err := db.Exec(ctxDB, `DELETE FROM pr_reviewers WHERE pr_id = $1`, prID)
	require.NoError(t, err)
	addReviewerDirect(t, ctxDB, db, prID, reviewerToReplace)

	t.Run("Reassign_Success", func(t *testing.T) {
		reassignReq := struct {
			PullRequestID uuid.UUID `json:"pull_request_id"`
			UserID        uuid.UUID `json:"old_user_id"`
		}{PullRequestID: prID, UserID: reviewerToReplace}
		body, _ := json.Marshal(reassignReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/reassign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			PR struct {
				Reviewers []uuid.UUID `json:"assigned_reviewers"`
			} `json:"pr"`
			ReplacedBy uuid.UUID `json:"replaced_by"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		newReviewerID := resp.ReplacedBy

		assert.NotEqual(t, reviewerToReplace, newReviewerID, "Старый и новый ревьюеры должны быть разными")
		assert.Contains(t, resp.PR.Reviewers, newReviewerID, "Новый ревьювер должен быть в списке")
		assert.NotContains(t, resp.PR.Reviewers, reviewerToReplace, "Старый ревьювер должен быть удален")
	})

	t.Run("Reassign_Conflict_NotAssigned", func(t *testing.T) {
		reassignReq := struct {
			PullRequestID uuid.UUID `json:"pull_request_id"`
			UserID        uuid.UUID `json:"old_user_id"`
		}{PullRequestID: prID, UserID: authorID}
		body, _ := json.Marshal(reassignReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/pullRequest/reassign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		var errResp struct {
			Error struct {
				Code string `json:"code"`
			}
		}
		json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "NOT_ASSIGNED", errResp.Error.Code)
	})

	db.Close()
}

func TestGetReviewAssignedPRs(t *testing.T) {
	r, db := SetupRouterForTesting(t)
	ctx := context.Background()

	const teamName = "getreview_squad"
	targetReviewerID := uuid.New()
	otherReviewerID := uuid.New()
	authorID := uuid.New()

	prTargetID1 := uuid.New()
	prTargetID2 := uuid.New()
	prOtherID3 := uuid.New()

	teamReq := TeamWithMembers{
		TeamName: teamName,
		Members: []struct {
			ID       uuid.UUID `json:"user_id"`
			Username string    `json:"username"`
			IsActive bool      `json:"is_active"`
		}{
			{ID: authorID, Username: "Author", IsActive: true},
			{ID: targetReviewerID, Username: "TargetRev", IsActive: true},
			{ID: otherReviewerID, Username: "OtherRev", IsActive: true},
		},
	}
	createTeamHTTP(t, r, teamReq)

	_ = createPRHTTP(t, r, prTargetID1, "TargetPR1", authorID)
	_ = createPRHTTP(t, r, prTargetID2, "TargetPR2", authorID)
	_ = createPRHTTP(t, r, prOtherID3, "OtherPR3", authorID)

	ctxDB := context.Background()
	_, err := db.Exec(ctxDB, `DELETE FROM pr_reviewers`)
	require.NoError(t, err)

	addReviewerDirect(t, ctxDB, db, prTargetID1, targetReviewerID)
	addReviewerDirect(t, ctxDB, db, prTargetID2, targetReviewerID)
	addReviewerDirect(t, ctxDB, db, prOtherID3, otherReviewerID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, "GET", "/api/users/getReview?user_id="+targetReviewerID.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		UserID       uuid.UUID `json:"user_id"`
		PullRequests []struct {
			ID       uuid.UUID `json:"pull_request_id"`
			Name     string    `json:"pull_request_name"`
			AuthorID uuid.UUID `json:"author_id"`
			Status   string    `json:"status"`
		} `json:"pull_requests"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, targetReviewerID, resp.UserID)
	assert.Len(t, resp.PullRequests, 2)

	ids := map[uuid.UUID]bool{}
	for _, p := range resp.PullRequests {
		ids[p.ID] = true
	}
	assert.True(t, ids[prTargetID1])
	assert.True(t, ids[prTargetID2])

	db.Close()
}
