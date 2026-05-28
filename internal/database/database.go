package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewDatabase(ctx context.Context, uri, dbName string) (*Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	db := &Database{
		Client: client,
		DB:     client.Database(dbName),
	}

	if err := db.EnsureIndexes(connectCtx); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return db, nil
}

func (d *Database) Close(ctx context.Context) error {
	if d == nil || d.Client == nil {
		return nil
	}

	return d.Client.Disconnect(ctx)
}

func (d *Database) EnsureIndexes(ctx context.Context) error {
	indexes := map[string][]mongo.IndexModel{
		"users": {
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("ux_users_email"),
			},
			{
				Keys:    bson.D{{Key: "yandex_id", Value: 1}},
				Options: options.Index().SetUnique(true).SetSparse(true).SetName("ux_users_yandex_id"),
			},
			{
				Keys:    bson.D{{Key: "vk_id", Value: 1}},
				Options: options.Index().SetUnique(true).SetSparse(true).SetName("ux_users_vk_id"),
			},
			{
				Keys:    bson.D{{Key: "reset_token_expires_at", Value: 1}},
				Options: options.Index().SetSparse(true).SetName("ix_users_reset_token_expires_at"),
			},
		},
		"tokens": {
			{
				Keys:    bson.D{{Key: "token_hash", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("ux_tokens_token_hash"),
			},
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "type", Value: 1}, {Key: "revoked", Value: 1}, {Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("ix_tokens_active_by_user"),
			},
		},
		"notes": {
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "deleted_at", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("ix_notes_user_deleted_created"),
			},
		},
		"files": {
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "deleted_at", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("ix_files_user_deleted_created"),
			},
			{
				Keys:    bson.D{{Key: "object_key", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("ux_files_object_key"),
			},
		},
	}

	for collection, models := range indexes {
		if _, err := d.DB.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("failed to create %s indexes: %w", collection, err)
		}
	}

	return nil
}
