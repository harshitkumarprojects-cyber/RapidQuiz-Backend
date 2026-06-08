package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/controllers"
	"github.com/harshitkumar7525/RapidQuiz/backend/middlewares"
)

func RegisterQuizRoutes(server *gin.Engine) {
	server.GET("/quizzes/:quizId", controllers.GetQuizByID)
	quizGrp := server.Group("/quizzes", middlewares.AuthMiddleware)
	quizGrp.POST("/", controllers.CreateQuiz)
	quizGrp.GET("/", controllers.GetQuizzes)
	quizGrp.PATCH("/:quizId", controllers.UpdateQuiz)
	quizGrp.DELETE("/:quizId", controllers.DeleteQuiz)
}
