package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/utils"
)

func AuthMiddleware(context *gin.Context) {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "authorization header is required",
		})
		return
	}
	token := strings.TrimPrefix(tokenString, "Bearer ")
	userId, err := utils.ValidateJWT(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}
	context.Set("userId", userId)
	context.Next()
}
