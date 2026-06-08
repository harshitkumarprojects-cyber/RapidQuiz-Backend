package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/database"
	"github.com/harshitkumar7525/RapidQuiz/backend/models"
	"github.com/harshitkumar7525/RapidQuiz/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterHandler(c *gin.Context) {
	var userData struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if userData.Name == "" || userData.Email == "" || userData.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "all fields are required",
		})
		return
	}

	var existingUser models.User

	err := database.Collection("users").
		Find(context.Background(), bson.M{
			"email": userData.Email,
		}).
		One(&existingUser)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "email already registered",
		})
		return
	}

	hashedPassword := utils.HashPassword(userData.Password)

	user := models.User{
		ID:        primitive.NewObjectID(),
		Name:      userData.Name,
		Email:     userData.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = database.Collection("users").
		InsertOne(context.Background(), user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to register user",
		})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.ID.Hex())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"token":   jwtToken,
	})
}

func LoginHandler(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if loginData.Email == "" || loginData.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email and password are required",
		})
		return
	}

	var user models.User

	err := database.Collection("users").
		Find(context.Background(), bson.M{
			"email": loginData.Email,
		}).
		One(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	if !utils.CheckPasswordHash(loginData.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.ID.Hex())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   jwtToken,
	})
}
