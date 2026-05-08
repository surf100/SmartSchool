package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	chatpb "chat-api/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MessageRepository struct {
	collection       *mongo.Collection
	groupsCollection *mongo.Collection
}

type Group struct {
	ID        string            `bson:"_id" json:"id"`
	GroupName string            `bson:"group_name" json:"group_name"`
	CreatorID string            `bson:"creator_id" json:"creator_id"`
	Members   []string          `bson:"members" json:"members"`
	Roles     map[string]string `bson:"roles" json:"roles"`
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection:       db.Collection("messages"),
		groupsCollection: db.Collection("groups"),
	}
}

type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	SenderID    string             `bson:"sender_id"`
	ReceiverID  string             `bson:"receiver_id"`
	Content     string             `bson:"content"`
	IsGroup     bool               `bson:"is_group"`
	GroupID     string             `bson:"chat_id"`
	CreatedAt   time.Time          `bson:"sent_at"`
	DeliveredAt string             `bson:"delivered_at"`
	IsDelivered bool               `bson:"is_delivered"`
	FileURL     string             `bson:"file_url"`
	FileType    string             `bson:"file_type"`
	FileName    string             `bson:"file_name"`
}

func (r *MessageRepository) Save(ctx context.Context, msg *chatpb.ChatMessage) error {
	sentAt, _ := time.Parse(time.RFC3339, msg.SentAt)
	deliveredAt := ""
	if msg.DeliveredAt != "" {
		if t, err := time.Parse(time.RFC3339, msg.DeliveredAt); err == nil {
			deliveredAt = t.Format(time.RFC3339)
		}
	}

	doc := bson.M{
		"chat_id":      msg.ChatId,
		"sender_id":    msg.SenderId,
		"receiver_id":  msg.ReceiverId,
		"content":      msg.Content,
		"sent_at":      sentAt,
		"delivered_at": deliveredAt,
		"is_delivered": msg.IsDelivered,
		"file_url":     msg.FileUrl,
		"file_type":    msg.FileType,
		"file_name":    msg.FileName,
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *MessageRepository) SaveMessage(ctx context.Context, senderID, receiverID, content string, isGroup bool, groupID string) error {
	msg := Message{
		SenderID:    senderID,
		ReceiverID:  receiverID,
		Content:     content,
		IsGroup:     isGroup,
		GroupID:     groupID,
		CreatedAt:   time.Now(),
		IsDelivered: false,
		DeliveredAt: "",
		FileURL:     "",
		FileType:    "",
		FileName:    "",
	}

	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

func (r *MessageRepository) GetHistory(ctx context.Context, chatID string, limit, offset int32) ([]*chatpb.ChatMessage, error) {
	filter := bson.M{"chat_id": chatID}

	findOptions := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "sent_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*chatpb.ChatMessage
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		var sentAtStr string
		if sentAtRaw, ok := doc["sent_at"]; ok {
			switch v := sentAtRaw.(type) {
			case primitive.DateTime:
				sentAtStr = v.Time().Format(time.RFC3339)
			case time.Time:
				sentAtStr = v.Format(time.RFC3339)
			case string:
				sentAtStr = v
			default:
				sentAtStr = ""
			}
		}

		var deliveredAtStr string
		if delivAtRaw, ok := doc["delivered_at"]; ok {
			switch v := delivAtRaw.(type) {
			case primitive.DateTime:
				deliveredAtStr = v.Time().Format(time.RFC3339)
			case time.Time:
				deliveredAtStr = v.Format(time.RFC3339)
			case string:
				deliveredAtStr = v
			default:
				deliveredAtStr = ""
			}
		}

		msg := &chatpb.ChatMessage{
			ChatId:      getString(doc["chat_id"]),
			SenderId:    getString(doc["sender_id"]),
			ReceiverId:  getString(doc["receiver_id"]),
			Content:     getString(doc["content"]),
			SentAt:      sentAtStr,
			DeliveredAt: deliveredAtStr,
			IsDelivered: getBool(doc["is_delivered"]),
			FileUrl:     getString(doc["file_url"]),
			FileType:    getString(doc["file_type"]),
			FileName:    getString(doc["file_name"]),
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MessageRepository) GetUndeliveredMessages(ctx context.Context, receiverID string) ([]*chatpb.ChatMessage, error) {
	filter := bson.M{
		"receiver_id":  receiverID,
		"is_delivered": false,
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*chatpb.ChatMessage
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		var sentAtStr string
		if sentAtRaw, ok := doc["sent_at"]; ok {
			if sentAtDT, ok := sentAtRaw.(primitive.DateTime); ok {
				sentAtStr = sentAtDT.Time().Format(time.RFC3339)
			} else if t, ok := sentAtRaw.(time.Time); ok {
				sentAtStr = t.Format(time.RFC3339)
			} else {
				log.Println("sent_at has unexpected type:", reflect.TypeOf(sentAtRaw))
				sentAtStr = ""
			}
		}

		msg := &chatpb.ChatMessage{
			ChatId:      doc["chat_id"].(string),
			SenderId:    doc["sender_id"].(string),
			ReceiverId:  doc["receiver_id"].(string),
			Content:     doc["content"].(string),
			SentAt:      sentAtStr,
			IsDelivered: false,
		}

		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *MessageRepository) MarkMessagesAsDelivered(ctx context.Context, receiverID string) error {
	filter := bson.M{
		"receiver_id":  receiverID,
		"is_delivered": false,
	}
	update := bson.M{
		"$set": bson.M{
			"is_delivered": true,
			"delivered_at": time.Now().Format(time.RFC3339),
		},
	}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *MessageRepository) UpdateMessageByID(id string, newContent string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"content": newContent}}
	_, err = r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *MessageRepository) DeleteMessageByID(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(context.Background(), filter)
	return err
}

func (r *MessageRepository) CreateGroup(ctx context.Context, id, name, creator string, members []string) error {
	group := bson.M{
		"_id":        id,
		"group_name": name,
		"creator_id": creator,
		"members":    members,
		"created_at": time.Now(),
	}
	_, err := r.groupsCollection.InsertOne(ctx, group)
	return err
}

func (r *MessageRepository) AddMember(ctx context.Context, groupID, userID string) error {
	filter := bson.M{
		"_id": groupID,
	}

	update := bson.M{
		"$addToSet": bson.M{
			"members": userID,
		},
	}
	result, err := r.groupsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		log.Println("Группа не найдена или пользователь уже есть в списке участников")
		return fmt.Errorf("group not found or user already in group")
	}

	return nil
}

func (r *MessageRepository) RemoveGroupMember(ctx context.Context, groupID string, userID string) error {
	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$pull":  bson.M{"members": userID},
		"$unset": bson.M{"roles." + userID: ""},
	}
	result, err := r.groupsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("user not in group or group not found")
	}
	return nil
}

func (r *MessageRepository) GetGroupMembers(ctx context.Context, groupID string) ([]string, error) {
	var group struct {
		Members []string `bson:"members"`
	}

	err := r.groupsCollection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "group '%s' not found", groupID)
		}
		return nil, status.Errorf(codes.Internal, "failed to get group: %v", err)
	}

	return group.Members, nil
}

func (r *MessageRepository) GetCollection() *mongo.Collection {
	return r.collection
}

func getString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func getBool(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
