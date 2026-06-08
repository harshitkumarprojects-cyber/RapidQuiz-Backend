package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Quiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`

	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`

	Questions []Question `bson:"questions" json:"questions"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
