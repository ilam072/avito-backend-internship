package router

import (
	"github.com/gin-gonic/gin"
	pullrequestrest "github.com/ilam072/avito-backend-internship/internal/pullrequest/rest"
	teamrest "github.com/ilam072/avito-backend-internship/internal/team/rest"
	userrest "github.com/ilam072/avito-backend-internship/internal/user/rest"
)

func New(
	userHandler *userrest.UserHandler,
	teamHandler *teamrest.TeamHandler,
	prHandler *pullrequestrest.PullRequestHandler,
) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	api := engine.Group("/api")

	// team
	api.POST("/team/add", teamHandler.CreateTeam)
	api.GET("/team/get", userHandler.GetTeam) // query ?team_name=

	// users
	api.POST("/users/setIsActive", userHandler.SetUserIsActive)
	api.GET("/users/getReview", prHandler.GetReview) // query ?user_id=

	// pull request
	api.POST("/pullRequest/create", prHandler.CreatePullRequest)
	api.POST("/pullRequest/merge", prHandler.MergePullRequest)
	api.POST("/pullRequest/reassign", prHandler.Reassign)

	return engine
}
