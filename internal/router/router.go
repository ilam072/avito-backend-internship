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

	// team
	engine.POST("/team/add", teamHandler.CreateTeam)
	engine.GET("/team/get", userHandler.GetTeam) // query ?team_name=

	// users
	engine.POST("/users/setIsActive", userHandler.SetUserIsActive)
	engine.GET("/users/getReview", prHandler.GetReview) // query ?user_id=

	// pull request
	engine.POST("/pullRequest/create", prHandler.CreatePullRequest)
	engine.POST("/pullRequest/merge", prHandler.MergePullRequest)
	engine.POST("/pullRequest/reassign", prHandler.Reassign)

	return engine
}
