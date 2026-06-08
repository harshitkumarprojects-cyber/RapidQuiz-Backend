package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshitkumar7525/RapidQuiz/backend/database"
	"github.com/harshitkumar7525/RapidQuiz/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/redis/go-redis/v9"
)

func leaderboardKey(gameID string) string {
	return fmt.Sprintf("leaderboard:%s", gameID)
}

func SubmitAnswer(c *gin.Context) {
	gameIDHex := c.Param("gameId")

	gameObjID, err := primitive.ObjectIDFromHex(gameIDHex)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid game id"})
		return
	}

	var request struct {
		ParticipantID string `json:"participant_id" binding:"required"`
		QuestionIndex int    `json:"question_index"`
		Answer        string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	participantObjID, err := primitive.ObjectIDFromHex(request.ParticipantID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid participant id"})
		return
	}

	ctx := context.Background()

	var game models.GameSession
	if err := database.Collection("game_sessions").
		Find(ctx, bson.M{"_id": gameObjID}).
		One(&game); err != nil {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}

	if game.Status != models.Running {
		c.JSON(400, gin.H{"error": "game is not currently running"})
		return
	}

	var quiz models.Quiz
	if err := database.Collection("quizzes").
		Find(ctx, bson.M{"_id": game.QuizID}).
		One(&quiz); err != nil {
		c.JSON(404, gin.H{"error": "quiz not found"})
		return
	}

	if request.QuestionIndex < 0 || request.QuestionIndex >= len(quiz.Questions) {
		c.JSON(400, gin.H{"error": "invalid question index"})
		return
	}

	var existing models.Answer
	dupErr := database.Collection("answers").
		Find(ctx, bson.M{
			"game_id":        gameObjID,
			"participant_id": participantObjID,
			"question_index": request.QuestionIndex,
		}).One(&existing)

	if dupErr == nil {
		c.JSON(409, gin.H{"error": "answer already submitted for this question"})
		return
	}

	question := quiz.Questions[request.QuestionIndex]
	isCorrect := question.CorrectAnswer == request.Answer

	score := 0
	if isCorrect {
		timeLimit := question.TimeLimit
		if timeLimit <= 0 {
			timeLimit = 30
		}

		// Calculate how many seconds the player took to answer.
		// QuestionStartedAt is stamped when the host starts the game or advances
		// to this question. If it's missing (old sessions), fall back to full score.
		elapsedSeconds := 0.0
		if game.QuestionStartedAt != nil {
			elapsedSeconds = time.Since(*game.QuestionStartedAt).Seconds()
			if elapsedSeconds < 0 {
				elapsedSeconds = 0
			}
		}

		// Scoring formula:
		//   Base score:  100 points for a correct answer
		//   Time bonus:  up to 100 extra points, scaling linearly from full marks
		//                (answered instantly) down to 0 (answered at the last second)
		//
		//   timeBonus = 100 * max(0, (timeLimit - elapsed) / timeLimit)
		//
		// Example with a 30-second question:
		//   answered in  1s → timeBonus = 97  → total = 197
		//   answered in 15s → timeBonus = 50  → total = 150
		//   answered in 29s → timeBonus =  3  → total = 103
		//   answered in 30s → timeBonus =  0  → total = 100
		remaining := float64(timeLimit) - elapsedSeconds
		if remaining < 0 {
			remaining = 0
		}
		timeBonus := int(100.0 * remaining / float64(timeLimit))
		score = 100 + timeBonus
	}

	answer := models.Answer{
		ID:            primitive.NewObjectID(),
		GameID:        gameObjID,
		ParticipantID: participantObjID,
		QuestionIndex: request.QuestionIndex,
		Answer:        request.Answer,
		IsCorrect:     isCorrect,
		Score:         score,
		AnsweredAt:    time.Now(),
	}

	if _, err := database.Collection("answers").InsertOne(ctx, answer); err != nil {
		c.JSON(500, gin.H{"error": "failed to save answer"})
		return
	}

	rKey := leaderboardKey(gameIDHex)
	if err := database.RedisClient.ZIncrBy(ctx, rKey, float64(score), request.ParticipantID).Err(); err != nil {
		// Non-fatal: log but still return success
		c.JSON(500, gin.H{"error": "failed to update leaderboard"})
		return
	}

	database.RedisClient.Expire(ctx, rKey, 24*time.Hour)

	c.JSON(200, gin.H{
		"is_correct": isCorrect,
		"score":      score,
		"message": func() string {
			if isCorrect {
				return "correct answer!"
			}
			return "wrong answer"
		}(),
	})
}

func GetLeaderboard(c *gin.Context) {
	gameIDHex := c.Param("gameId")

	if _, err := primitive.ObjectIDFromHex(gameIDHex); err != nil {
		c.JSON(400, gin.H{"error": "invalid game id"})
		return
	}

	ctx := context.Background()
	rKey := leaderboardKey(gameIDHex)

	results, err := database.RedisClient.ZRevRangeWithScores(ctx, rKey, 0, 19).Result()
	if err != nil && err != redis.Nil {
		c.JSON(500, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	type LeaderboardEntry struct {
		Rank          int     `json:"rank"`
		ParticipantID string  `json:"participant_id"`
		Name          string  `json:"name"`
		Score         float64 `json:"score"`
	}

	entries := make([]LeaderboardEntry, 0, len(results))

	for i, z := range results {
		participantIDHex := z.Member.(string)
		participantObjID, err := primitive.ObjectIDFromHex(participantIDHex)

		name := "Unknown"
		if err == nil {
			var participant models.Participant
			if err := database.Collection("participants").
				Find(ctx, bson.M{"_id": participantObjID}).
				One(&participant); err == nil {
				name = participant.Name
			}
		}

		entries = append(entries, LeaderboardEntry{
			Rank:          i + 1,
			ParticipantID: participantIDHex,
			Name:          name,
			Score:         z.Score,
		})
	}

	c.JSON(200, gin.H{
		"game_id":     gameIDHex,
		"leaderboard": entries,
	})
}
