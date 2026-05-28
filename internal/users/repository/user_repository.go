package repository

import (
	"context"
	"time"

	"Web_lab/internal/database"
	"Web_lab/internal/users/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{collection: db.DB.Collection("users")}
}

func (r *UserRepository) Create(user *models.User) error {
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(context.Background(), user)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	return r.findOne(activeUserFilter(bson.D{{Key: "email", Value: email}}))
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	return r.findOne(activeUserFilter(bson.D{{Key: "_id", Value: id}}))
}

func (r *UserRepository) FindByYandexID(yandexID string) (*models.User, error) {
	return r.findOne(activeUserFilter(bson.D{{Key: "yandex_id", Value: yandexID}}))
}

func (r *UserRepository) FindUsersWithActiveResetToken() ([]models.User, error) {
	filter := activeUserFilter(bson.D{
		{Key: "reset_token_hash", Value: bson.D{{Key: "$exists", Value: true}}},
		{Key: "reset_token_expires_at", Value: bson.D{{Key: "$gt", Value: time.Now().UTC()}}},
	})

	cursor, err := r.collection.Find(context.Background(), filter, options.Find())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now().UTC()

	_, err := r.collection.ReplaceOne(
		context.Background(),
		activeUserFilter(bson.D{{Key: "_id", Value: user.ID}}),
		user,
	)
	return err
}

func (r *UserRepository) UpdateProfile(user *models.User) error {
	return r.Update(user)
}

func (r *UserRepository) findOne(filter bson.D) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func activeUserFilter(base bson.D) bson.D {
	filter := make(bson.D, 0, len(base)+1)
	filter = append(filter, base...)
	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	return filter
}
