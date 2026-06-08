package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameStatus string

const (
	Waiting GameStatus = "waiting"
	Running GameStatus = "running"
	Paused  GameStatus = "paused"
	Ended   GameStatus = "ended"
)

type GameSession struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	QuizID   primitive.ObjectID `bson:"quiz_id"`
	HostID   primitive.ObjectID `bson:"host_id"`
	RoomCode string             `bson:"room_code"`

	Status GameStatus `bson:"status"`

	CurrentQuestion int `bson:"current_question"`
	QuestionStartedAt *time.Time `bson:"question_started_at,omitempty"`

	StartedAt *time.Time `bson:"started_at,omitempty"`
	EndedAt   *time.Time `bson:"ended_at,omitempty"`
}