package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/controllers"
)

func RegisterAuthRoutes(server *gin.Engine) {
	authGroup := server.Group("/auth")
	authGroup.POST("/register", controllers.RegisterHandler)
	authGroup.POST("/login", controllers.LoginHandler)
}
