package main

import (
	"context"
	"github.com/ilam072/avito-backend-internship/internal/config"
	prrepo "github.com/ilam072/avito-backend-internship/internal/pullrequest/repo"
	pullrequestrest "github.com/ilam072/avito-backend-internship/internal/pullrequest/rest"
	pullrequestservice "github.com/ilam072/avito-backend-internship/internal/pullrequest/service"
	"github.com/ilam072/avito-backend-internship/internal/router"
	teamrepo "github.com/ilam072/avito-backend-internship/internal/team/repo"
	teamrest "github.com/ilam072/avito-backend-internship/internal/team/rest"
	teamservice "github.com/ilam072/avito-backend-internship/internal/team/service"
	userrepo "github.com/ilam072/avito-backend-internship/internal/user/repo"
	userrest "github.com/ilam072/avito-backend-internship/internal/user/rest"
	userservice "github.com/ilam072/avito-backend-internship/internal/user/service"
	"github.com/ilam072/avito-backend-internship/internal/validator"
	"github.com/ilam072/avito-backend-internship/pkg/db"
	"github.com/rs/zerolog/log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize config
	cfg := config.MustLoad()

	// Connect to DB
	DB, err := db.OpenDB(ctx, cfg.DB)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Initialize validator
	v := validator.New()

	// Initialize user, team and pull request repositories
	userRepo := userrepo.New(DB)
	teamRepo := teamrepo.New(DB)
	prRepo := prrepo.New(DB)

	// Initialize user, team and pull request services
	user := userservice.NewUser(userRepo, teamRepo)
	item := teamservice.NewTeam(teamRepo)
	pullRequest := pullrequestservice.NewPullRequest(userRepo, prRepo)

	// Initialize user, team and pull request handlers
	userHandler := userrest.NewUserHandler(user, v)
	teamHandler := teamrest.NewTeamHandler(item, v)
	prHandler := pullrequestrest.NewPullRequestHandler(pullRequest, v)

	// Initialize Gin engine and set routes
	engine := router.New(userHandler, teamHandler, prHandler)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(withTimeout); err != nil {
		log.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	DB.Close()
}
