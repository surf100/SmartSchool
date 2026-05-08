package migrations

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RunMongoMigrations(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	db := client.Database("chatDB")

	//  Коллекция messages
	messages := db.Collection("messages")

	_, err = messages.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "chat_id", Value: 1}}},
		{Keys: bson.D{{Key: "sender_id", Value: 1}}},
		{Keys: bson.D{{Key: "receiver_id", Value: 1}}},
		{Keys: bson.D{{Key: "sent_at", Value: 1}}},
	})
	if err != nil {
		return err
	}

	// (опционально) Вставка тестового сообщения
	_, _ = messages.InsertOne(ctx, bson.M{
		"sent_at":      time.Date(2025, 8, 2, 12, 14, 0, 0, time.UTC),
		"delivered_at": "2025-08-02T12:15:00Z",
		"is_delivered": true,
		"file_url":     "/static/uploads/example.jpg",
		"chat_id":      "demo_group_8",
		"sender_id":    "u3",
		"receiver_id":  "u101",
		"content":      "elit consectetur irure",
		"file_name":    "Excepteur",
		"file_type":    "image",
	})

	// Коллекция groups
	groups := db.Collection("groups")

	_, err = groups.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "creator_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	})
	if err != nil {
		return err
	}

	// (опционально) Вставка тестовой группы
	_, _ = groups.InsertOne(ctx, bson.M{
		"_id":        "demo_group_8",
		"group_name": "Demo",
		"creator_id": "u1",
		"members":    []string{"u3", "u5", "u7", "u101"},
		"created_at": time.Date(2025, 8, 2, 7, 13, 7, 154_000_000, time.UTC),
	})

	log.Println("MongoDB migrations applied successfully.")
	return nil
}
