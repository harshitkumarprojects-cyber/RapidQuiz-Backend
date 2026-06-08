package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/database"
	"github.com/harshitkumar7525/RapidQuiz/backend/models"
	"github.com/harshitkumar7525/RapidQuiz/backend/utils"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateQuiz(c *gin.Context) {
	userId := utils.GetPrimitiveUserID(c)
	var quiz models.Quiz

	if err := c.ShouldBindJSON(&quiz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if quiz.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title is required",
		})
		return
	}

	if len(quiz.Questions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "quiz must contain at least one question",
		})
		return
	}

	for i, q := range quiz.Questions {
		if err := q.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("question %d: %s", i+1, err.Error()),
			})
			return
		}
	}

	quiz.ID = primitive.NewObjectID()
	quiz.CreatedAt = time.Now()
	quiz.UpdatedAt = time.Now()
	quiz.CreatedBy = userId

	_, err := database.Collection("quizzes").
		InsertOne(context.Background(), quiz)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create quiz",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "quiz created successfully",
	})
}

func GetQuizzes(c *gin.Context) {
	userID := utils.GetPrimitiveUserID(c)
	var quizzes []models.Quiz

	err := database.Collection("quizzes").
		Find(context.Background(), bson.M{
			"created_by": userID,
		}).
		All(&quizzes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}

func UpdateQuiz(c *gin.Context) {
	quizID := c.Param("quizId")
	userId := utils.GetPrimitiveUserID(c)

	quizObjID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid quiz ID",
		})
		return
	}

	var quiz models.Quiz

	err = database.Collection("quizzes").
		Find(context.Background(), bson.M{
			"_id":        quizObjID,
			"created_by": userId,
		}).
		One(&quiz)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "quiz not found or you do not have permission to update it",
		})
		return
	}

	var updateData models.Quiz

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for i, q := range updateData.Questions {
		if err := q.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("question %d: %s", i+1, err.Error()),
			})
			return
		}
	}

	update := bson.M{
		"$set": bson.M{
			"title":       updateData.Title,
			"description": updateData.Description,
			"questions":   updateData.Questions,
			"updated_at":  time.Now(),
		},
	}

	err = database.Collection("quizzes").
		UpdateOne(context.Background(), bson.M{
			"_id":        quizObjID,
			"created_by": userId,
		}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update quiz",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "quiz updated successfully",
	})
}

func DeleteQuiz(c *gin.Context) {
	quizID := c.Param("quizId")
	userId := utils.GetPrimitiveUserID(c)

	quizObjID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid quiz ID",
		})
		return
	}

	ctx := context.Background()

	// First verify the quiz exists and belongs to this user.
	// qmgo's Remove() silently succeeds even when no document matches,
	// so we must check ownership explicitly before deleting.
	var quiz models.Quiz
	err = database.Collection("quizzes").
		Find(ctx, bson.M{"_id": quizObjID}).
		One(&quiz)

	if err != nil {
		if err == qmgo.ErrNoSuchDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "quiz not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch quiz",
		})
		return
	}

	if quiz.CreatedBy != userId {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you do not have permission to delete this quiz",
		})
		return
	}

	err = database.Collection("quizzes").
		Remove(ctx, bson.M{"_id": quizObjID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete quiz",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "quiz deleted successfully",
	})
}

func GetQuizByID(c *gin.Context) {
	quizId := c.Param("quizId")

	quizObjID, err := primitive.ObjectIDFromHex(quizId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid quiz ID",
		})
		return
	}

	var quiz models.Quiz

	err = database.Collection("quizzes").
		Find(context.Background(), bson.M{
			"_id": quizObjID,
		}).
		One(&quiz)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "quiz not found",
		})
		return
	}

	c.JSON(http.StatusOK, quiz)
}
