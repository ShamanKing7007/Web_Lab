package repository

import (
	"context"
	"time"

	"Web_lab/internal/database"
	"Web_lab/internal/storage/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type FileRepository interface {
	Create(file *models.File) error
	FindByID(id uuid.UUID) (*models.File, error)
	Delete(id uuid.UUID) (bool, error)
}

type FileRepositoryImpl struct {
	collection *mongo.Collection
}

func NewFileRepository(db *database.Database) FileRepository {
	return &FileRepositoryImpl{collection: db.DB.Collection("files")}
}

func (r *FileRepositoryImpl) Create(file *models.File) error {
	now := time.Now().UTC()
	if file.CreatedAt.IsZero() {
		file.CreatedAt = now
	}
	file.UpdatedAt = now

	_, err := r.collection.InsertOne(context.Background(), file)
	return err
}

func (r *FileRepositoryImpl) FindByID(id uuid.UUID) (*models.File, error) {
	var file models.File
	err := r.collection.FindOne(context.Background(), activeFilter(bson.D{{Key: "_id", Value: id}})).Decode(&file)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *FileRepositoryImpl) Delete(id uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
		context.Background(),
		activeFilter(bson.D{{Key: "_id", Value: id}}),
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "deleted_at", Value: now},
			{Key: "updated_at", Value: now},
		}}},
	)
	if err != nil {
		return false, err
	}

	return result.MatchedCount > 0, nil
}

func activeFilter(base bson.D) bson.D {
	filter := make(bson.D, 0, len(base)+1)
	filter = append(filter, base...)
	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	return filter
}
