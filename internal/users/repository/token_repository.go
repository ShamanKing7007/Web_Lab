package repository

import (
	"context"
	"time"

	"Web_lab/internal/database"
	"Web_lab/internal/users/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TokenRepository struct {
	collection *mongo.Collection
}

func NewTokenRepository(db *database.Database) *TokenRepository {
	return &TokenRepository{collection: db.DB.Collection("tokens")}
}

func (r *TokenRepository) Create(token *models.Token) error {
	now := time.Now().UTC()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	token.UpdatedAt = now

	_, err := r.collection.InsertOne(context.Background(), token)
	return err
}

func (r *TokenRepository) FindByHash(tokenHash string) (*models.Token, error) {
	return r.findOne(activeTokenFilter(bson.D{
		{Key: "token_hash", Value: tokenHash},
		{Key: "revoked", Value: false},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: time.Now().UTC()}}},
	}))
}

func (r *TokenRepository) FindActiveByUser(userID uuid.UUID, tokenType string) ([]models.Token, error) {
	filter := activeTokenFilter(bson.D{
		{Key: "user_id", Value: userID},
		{Key: "type", Value: tokenType},
		{Key: "revoked", Value: false},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: time.Now().UTC()}}},
	})

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tokens []models.Token
	if err := cursor.All(context.Background(), &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *TokenRepository) RevokeByID(tokenID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.collection.UpdateOne(
		context.Background(),
		activeTokenFilter(bson.D{{Key: "_id", Value: tokenID}}),
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "revoked", Value: true},
			{Key: "updated_at", Value: now},
		}}},
	)
	return err
}

func (r *TokenRepository) RevokeAll(userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.collection.UpdateMany(
		context.Background(),
		activeTokenFilter(bson.D{
			{Key: "user_id", Value: userID},
			{Key: "revoked", Value: false},
		}),
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "revoked", Value: true},
			{Key: "updated_at", Value: now},
		}}},
	)
	return err
}

func (r *TokenRepository) findOne(filter bson.D) (*models.Token, error) {
	var token models.Token
	err := r.collection.FindOne(context.Background(), filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func activeTokenFilter(base bson.D) bson.D {
	filter := make(bson.D, 0, len(base)+1)
	filter = append(filter, base...)
	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	return filter
}
