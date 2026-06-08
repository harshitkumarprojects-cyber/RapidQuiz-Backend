package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Answer struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GameID        primitive.ObjectID `bson:"game_id" json:"game_id"`
	ParticipantID primitive.ObjectID `bson:"participant_id" json:"participant_id"`
	QuestionIndex int                `bson:"question_index" json:"question_index"`
	Answer        string             `bson:"answer" json:"answer"`
	IsCorrect     bool               `bson:"is_correct" json:"is_correct"`
	Score         int                `bson:"score" json:"score"`
	AnsweredAt    time.Time          `bson:"answered_at" json:"answered_at"`
}