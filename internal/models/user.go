package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"password"`
	SecretKey string             `bson:"secret_key" json:"secret_key"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
}
