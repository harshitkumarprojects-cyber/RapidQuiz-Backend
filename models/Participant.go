package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Participant struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GameID   primitive.ObjectID `bson:"game_id" json:"game_id"`
	Name     string             `bson:"name" json:"name"`
	JoinedAt time.Time          `bson:"joined_at" json:"joined_at"`
}
