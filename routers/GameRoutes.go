package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/controllers"
	"github.com/harshitkumar7525/RapidQuiz/backend/middlewares"
	"github.com/harshitkumar7525/RapidQuiz/backend/websocket"
)

func RegisterGameRoutes(server *gin.Engine) {
	server.POST("/games/create", middlewares.AuthMiddleware, controllers.StartGame)
	server.POST("/games/join", controllers.JoinGame)
	server.GET("/games/:gameId", middlewares.AuthMiddleware, controllers.GetGameByID)
	server.PATCH("/games/:gameId/status", middlewares.AuthMiddleware, controllers.UpdateGameStatus)
	server.POST("/games/:gameId/next", middlewares.AuthMiddleware, controllers.NextQuestion)
	server.GET("/ws/:roomCode", websocket.HandleWebSocket)
}
