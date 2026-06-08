package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/database"
	"github.com/harshitkumar7525/RapidQuiz/backend/routers"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	database.Connect()
	defer database.Disconnect()
	database.ConnectRedis()
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("CLIENT_URL")},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to RapidQuiz API",
		})
	})
	routers.RegisterAuthRoutes(server)
	routers.RegisterQuizRoutes(server)
	routers.RegisterGameRoutes(server)
	routers.RegisterLeaderboardRoutes(server)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
