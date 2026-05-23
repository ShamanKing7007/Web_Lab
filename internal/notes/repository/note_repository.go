package repository

import (
	"context"
	"time"

	"Web_lab/internal/database"
	"Web_lab/internal/notes/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NoteRepository interface {
	Create(note *models.Note) error
	FindByID(id uuid.UUID) (*models.Note, error)
	FindAll(offset, limit int) ([]models.Note, int64, error)
	FindAllByUser(userID uuid.UUID, offset, limit int) ([]models.Note, int64, error)
	Update(note *models.Note) error
	Delete(id uuid.UUID) (bool, error)
}

type NoteRepositoryImpl struct {
	collection *mongo.Collection
}

func NewNoteRepository(db *database.Database) NoteRepository {
	return &NoteRepositoryImpl{collection: db.DB.Collection("notes")}
}

func (r *NoteRepositoryImpl) Create(note *models.Note) error {
	now := time.Now().UTC()
	if note.CreatedAt.IsZero() {
		note.CreatedAt = now
	}
	note.UpdatedAt = now

	_, err := r.collection.InsertOne(context.Background(), note)
	return err
}

func (r *NoteRepositoryImpl) FindByID(id uuid.UUID) (*models.Note, error) {
	var note models.Note
	err := r.collection.FindOne(context.Background(), activeFilter(bson.D{{Key: "_id", Value: id}})).Decode(&note)
	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (r *NoteRepositoryImpl) FindAll(offset, limit int) ([]models.Note, int64, error) {
	filter := activeFilter(bson.D{})
	return r.findPage(filter, offset, limit)
}

func (r *NoteRepositoryImpl) FindAllByUser(userID uuid.UUID, offset, limit int) ([]models.Note, int64, error) {
	filter := activeFilter(bson.D{{Key: "user_id", Value: userID}})
	return r.findPage(filter, offset, limit)
}

func (r *NoteRepositoryImpl) Update(note *models.Note) error {
	note.UpdatedAt = time.Now().UTC()

	_, err := r.collection.ReplaceOne(
		context.Background(),
		activeFilter(bson.D{{Key: "_id", Value: note.ID}}),
		note,
	)
	return err
}

func (r *NoteRepositoryImpl) Delete(id uuid.UUID) (bool, error) {
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

func (r *NoteRepositoryImpl) findPage(filter bson.D, offset, limit int) ([]models.Note, int64, error) {
	ctx := context.Background()
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notes []models.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

func activeFilter(base bson.D) bson.D {
	filter := make(bson.D, 0, len(base)+1)
	filter = append(filter, base...)
	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	return filter
}
