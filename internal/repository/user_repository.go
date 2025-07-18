package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/fraiday-org/api-service/internal/models"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ValidateUser(ctx context.Context, username, password string) (*models.User, error) {
	user, err := r.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, errors.New("invalid credentials or inactive user")
	}
	if user.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

// For demo: Insert a user if not exists (idempotent)
func (r *UserRepository) EnsureDemoUser(ctx context.Context) error {
	username := "fraiday-dev-user"
	password := "fraiday-dev-pwd"
	secretKey := "test"
	createdAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	user, err := r.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user != nil {
		return nil
	}
	_, err = r.collection.InsertOne(ctx, bson.M{
		"username":   username,
		"password":   password,
		"secret_key": secretKey,
		"created_at": createdAt,
		"is_active":  true,
	})
	return err
}
