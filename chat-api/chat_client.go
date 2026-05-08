package main

import (
	"bufio"
	chatpb "chat-api/proto/gen"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
)

func connectMongo() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	return client.Database("chatdb").Collection("groups")
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := chatpb.NewChatServiceClient(conn)

	stream, err := client.SendMessage(context.Background())
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите ваш user_id: ")
	senderID, _ := reader.ReadString('\n')
	senderID = trim(senderID)

	fmt.Print("Вы хотите отправлять в группу? (yes/no): ")
	groupChoice, _ := reader.ReadString('\n')
	groupChoice = trim(groupChoice)

	receiverID := ""

	if groupChoice == "yes" {
		fmt.Print("Введите group_id: ")
		receiverID, _ = reader.ReadString('\n')
		receiverID = trim(receiverID)
	} else {
		fmt.Print("Введите receiver_id (user): ")
		receiverID, _ = reader.ReadString('\n')
		receiverID = trim(receiverID)
	}
	groupColl := connectMongo()

	if groupChoice == "yes" {
		groupDoc := groupColl.FindOne(context.Background(), map[string]interface{}{
			"_id": receiverID,
		})

		var group struct {
			ID      string   `bson:"_id"`
			Members []string `bson:"members"`
		}

		if err := groupDoc.Decode(&group); err != nil {
			log.Fatalf("Группа с ID %s не найдена: %v", receiverID, err)
		}

		found := false
		for _, member := range group.Members {
			if member == senderID {
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("Пользователь %s не состоит в группе %s", senderID, receiverID)
		}
	}

	var chatID string
	if groupChoice == "yes" {
		chatID = receiverID
	} else {
		chatID = fmt.Sprintf("%s_%s", senderID, receiverID)
	}

	regMsg := &chatpb.ChatMessage{
		ChatId:      chatID,
		SenderId:    senderID,
		ReceiverId:  receiverID,
		IsGroup:     groupChoice == "yes",
		Content:     "__register__",
		SentAt:      time.Now().Format(time.RFC3339),
		IsDelivered: false,
	}

	_ = stream.Send(regMsg)

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Printf("Ошибка при приёме: %v\n", err)
				break
			}

			if in.FileUrl != "" {
				fmt.Printf("\n[%s] %s прислал файл (%s): %s\n> ",
					in.ChatId, in.SenderId, in.FileType, in.FileUrl)
			} else {
				fmt.Printf("\n[%s] %s: %s\n> ",
					in.ChatId, in.SenderId, in.Content)
			}
		}
	}()
	fmt.Println("Для отправки файла напишите префикс /file и введите путь")
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = trim(text)

		if text == "" {
			continue
		}

		if strings.HasPrefix(text, "/file") {
			path := strings.TrimSpace(strings.TrimPrefix(text, "/file"))
			fileName := filepath.Base(path)
			fileType := filepath.Ext(path)
			fileURL := "/static/uploads/" + fileName

			inputFile, err := os.Open(path)
			if err != nil {
				log.Printf("Ошибка открытия файла: %v\n", err)
				continue
			}
			defer inputFile.Close()

			destPath := filepath.Join("static/uploads", fileName)
			outputFile, err := os.Create(destPath)
			if err != nil {
				log.Printf("Ошибка создания файла: %v\n", err)
				continue
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, inputFile)
			if err != nil {
				log.Printf("Ошибка копирования файла: %v\n", err)
				continue
			}

			msg := &chatpb.ChatMessage{
				ChatId:      chatID,
				SenderId:    senderID,
				ReceiverId:  receiverID,
				IsGroup:     groupChoice == "yes",
				Content:     fileURL,
				FileUrl:     fileURL,
				FileName:    fileName,
				FileType:    detectFileType(fileType),
				SentAt:      time.Now().Format(time.RFC3339),
				IsDelivered: false,
				DeliveredAt: "",
			}

			if err := stream.Send(msg); err != nil {
				log.Printf("Ошибка при отправке файла: %v\n", err)
			}

			continue
		}

		msg := &chatpb.ChatMessage{
			ChatId:      chatID,
			SenderId:    senderID,
			ReceiverId:  receiverID,
			Content:     text,
			IsGroup:     groupChoice == "yes",
			SentAt:      time.Now().Format(time.RFC3339),
			IsDelivered: false,
			DeliveredAt: "",
		}

		if err := stream.Send(msg); err != nil {
			log.Printf("Не удалось отправить: %v\n", err)
			break
		}
	}

}

func detectFileType(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3", ".wav":
		return "audio"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	case ".mp4", ".mov", ".avi":
		return "video"
	default:
		return "file"
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
