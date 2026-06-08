package controllers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/database"
	"github.com/harshitkumar7525/RapidQuiz/backend/models"
	"github.com/harshitkumar7525/RapidQuiz/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StartGame(c *gin.Context) {
	var request struct {
		QuizID string `json:"quiz_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := utils.GetPrimitiveUserID(c)

	quizObjID, err := primitive.ObjectIDFromHex(request.QuizID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid quiz id",
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
		c.JSON(404, gin.H{
			"error": "quiz not found",
		})
		return
	}

	if quiz.CreatedBy != userID {
		c.JSON(403, gin.H{
			"error": "you are not the creator of this quiz",
		})
		return
	}

	roomCode, err := utils.GenerateRoomCode()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to generate room code",
		})
		return
	}

	game := models.GameSession{
		ID:              primitive.NewObjectID(),
		QuizID:          quiz.ID,
		HostID:          userID,
		RoomCode:        roomCode,
		Status:          models.Waiting,
		CurrentQuestion: 0,
	}

	_, err = database.Collection("game_sessions").
		InsertOne(context.Background(), game)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to create game session",
		})
		return
	}

	c.JSON(201, gin.H{
		"message":         "game session created successfully",
		"room_code":       game.RoomCode,
		"game_id":         game.ID.Hex(),
		"status":          game.Status,
		"currentQuestion": game.CurrentQuestion,
	})
}

func GetGameByID(c *gin.Context) {
	gameIDHex := c.Param("gameId")

	gameObjID, err := primitive.ObjectIDFromHex(gameIDHex)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid game id"})
		return
	}

	var game models.GameSession
	if err := database.Collection("game_sessions").
		Find(context.Background(), bson.M{"_id": gameObjID}).
		One(&game); err != nil {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}

	c.JSON(200, gin.H{
		"game_id":          game.ID.Hex(),
		"quiz_id":          game.QuizID.Hex(),
		"room_code":        game.RoomCode,
		"status":           game.Status,
		"current_question": game.CurrentQuestion,
	})
}

func UpdateGameStatus(c *gin.Context) {
	gameIDHex := c.Param("gameId")
	userID := utils.GetPrimitiveUserID(c)

	gameObjID, err := primitive.ObjectIDFromHex(gameIDHex)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid game id"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	validStatuses := map[string]bool{
		"running": true,
		"paused":  true,
		"ended":   true,
	}
	if !validStatuses[body.Status] {
		c.JSON(400, gin.H{"error": "invalid status, must be one of: running, paused, ended"})
		return
	}

	now := time.Now()
	updateFields := bson.M{"status": body.Status}
	if body.Status == "running" {
		updateFields["started_at"] = now
		// Stamp when question 0 became active so elapsed time can be calculated
		updateFields["question_started_at"] = now
	} else if body.Status == "ended" {
		updateFields["ended_at"] = now
	}

	err = database.Collection("game_sessions").
		UpdateOne(context.Background(),
			bson.M{"_id": gameObjID, "host_id": userID},
			bson.M{"$set": updateFields},
		)
	if err != nil {
		c.JSON(404, gin.H{"error": "game not found or you are not the host"})
		return
	}

	c.JSON(200, gin.H{"message": "status updated", "status": body.Status})
}

func NextQuestion(c *gin.Context) {
	gameIDHex := c.Param("gameId")
	userID := utils.GetPrimitiveUserID(c)

	gameObjID, err := primitive.ObjectIDFromHex(gameIDHex)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid game id"})
		return
	}

	// Accept an explicit index from the host so skipping stays in sync.
	// Falls back to auto-increment if the body is absent or malformed.
	var body struct {
		Index *int `json:"index"`
	}
	_ = c.ShouldBindJSON(&body)

	var game models.GameSession
	if err := database.Collection("game_sessions").
		Find(context.Background(), bson.M{"_id": gameObjID, "host_id": userID}).
		One(&game); err != nil {
		c.JSON(404, gin.H{"error": "game not found or you are not the host"})
		return
	}

	nextIdx := game.CurrentQuestion + 1
	if body.Index != nil {
		nextIdx = *body.Index
	}

	questionStartedAt := time.Now()

	err = database.Collection("game_sessions").
		UpdateOne(context.Background(),
			bson.M{"_id": gameObjID, "host_id": userID},
			bson.M{"$set": bson.M{
				"current_question":    nextIdx,
				"question_started_at": questionStartedAt,
			}},
		)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to advance question"})
		return
	}

	c.JSON(200, gin.H{"message": "question advanced", "current_question": nextIdx})
}

func JoinGame(c *gin.Context) {
	var request struct {
		RoomCode string `json:"room_code" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var game models.GameSession

	err := database.Collection("game_sessions").
		Find(context.Background(), bson.M{
			"room_code": request.RoomCode,
			"status": bson.M{
				"$ne": models.Ended,
			},
		}).
		One(&game)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "room not found",
		})
		return
	}

	var existingParticipant models.Participant

	err = database.Collection("participants").
		Find(context.Background(), bson.M{
			"game_id": game.ID,
			"name":    request.Name,
		}).
		One(&existingParticipant)

	if err == nil {
		c.JSON(409, gin.H{
			"error": "name already taken in this room",
		})
		return
	}

	participant := models.Participant{
		ID:       primitive.NewObjectID(),
		GameID:   game.ID,
		Name:     request.Name,
		JoinedAt: time.Now(),
	}

	_, err = database.Collection("participants").
		InsertOne(context.Background(), participant)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to join game",
		})
		return
	}

	c.JSON(200, gin.H{
		"message":        "joined successfully",
		"participant_id": participant.ID.Hex(),
		"game_id":        game.ID.Hex(),
	})
}