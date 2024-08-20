package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Event struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description" bson:"description"`
	StartTime   time.Time          `json:"start_time" bson:"start_time"`
	EndTime     time.Time          `json:"end_time" bson:"end_time"`
}
