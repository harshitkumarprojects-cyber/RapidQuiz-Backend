package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/controllers"
)

func RegisterLeaderboardRoutes(server *gin.Engine) {
	server.POST("/games/:gameId/answer", controllers.SubmitAnswer)
	server.GET("/games/:gameId/leaderboard", controllers.GetLeaderboard)
}