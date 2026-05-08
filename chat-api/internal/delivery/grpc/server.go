package grpc

import (
	chatws "chat-api/internal/delivery/websocket"
	"chat-api/internal/repository"
	chatpb "chat-api/proto/gen"
	"context"
	"io"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	msgRepo     *repository.MessageRepository
	clients     map[string]chatpb.ChatService_SendMessageServer
	onlineUsers sync.Map
}

func NewChatServer(repo *repository.MessageRepository) *ChatServer {
	return &ChatServer{
		msgRepo:     repo,
		clients:     make(map[string]chatpb.ChatService_SendMessageServer),
		onlineUsers: sync.Map{},
	}
}

func (s *ChatServer) SendMessage(stream chatpb.ChatService_SendMessageServer) error {
	ctx := context.Background()
	var senderID string

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("%s отключился", senderID)
			delete(s.clients, senderID)
			return nil
		}
		if err != nil {
			log.Printf("Recv error: %v", err)
			delete(s.clients, senderID)
			return err
		}

		if msg.Content == "__register__" {
			senderID = msg.SenderId
			s.clients[senderID] = stream
			log.Printf("Зарегистрирован клиент %s", senderID)

			undelivered, err := s.msgRepo.GetUndeliveredMessages(ctx, senderID)
			if err == nil {
				for _, m := range undelivered {
					_ = stream.Send(m)
					time.Sleep(500 * time.Millisecond)
				}
				_ = s.msgRepo.MarkMessagesAsDelivered(ctx, senderID)
			}
			continue
		}

		if msg.IsGroup {
			groupID := msg.ReceiverId
			members, err := s.msgRepo.GetGroupMembers(ctx, groupID)
			if err != nil {
				log.Printf("Ошибка получения участников группы %s: %v", groupID, err)
				continue
			}

			log.Printf("Групповое сообщение от %s в %s", msg.SenderId, groupID)

			for _, userID := range members {
				if userID == msg.SenderId {
					continue
				}
				if userStream, ok := s.clients[userID]; ok {
					_ = userStream.Send(msg)
					log.Printf("Доставлено %s → %s", msg.SenderId, userID)
				} else {
					log.Printf("%s offline", userID)
				}
			}

			_ = s.msgRepo.Save(ctx, msg)
			continue
		}

		log.Printf("%s → %s: %s", msg.SenderId, msg.ReceiverId, msg.Content)
		_ = s.msgRepo.Save(ctx, msg)

		if receiverStream, ok := s.clients[msg.ReceiverId]; ok {
			if err := receiverStream.Send(msg); err != nil {
				log.Printf("Ошибка отправки пользователю %s: %v", msg.ReceiverId, err)
				return err
			}
		} else {
			log.Printf("%s is offline or not registered", msg.ReceiverId)
		}

	}
}

func (s *ChatServer) GetHistory(ctx context.Context, req *chatpb.ChatHistoryRequest) (*chatpb.ChatHistoryResponse, error) {
	messages, err := s.msgRepo.GetHistory(ctx, req.ChatId, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Ошибка при получении истории: %v", err)
		return nil, err
	}

	return &chatpb.ChatHistoryResponse{
		Messages: messages,
	}, nil
}

func (s *ChatServer) Init(ctx context.Context, req *chatpb.UserRequest) (*chatpb.StatusResponse, error) {
	log.Printf("User %s initialized", req.GetUserId())
	s.onlineUsers.Store(req.GetUserId(), true)

	return &chatpb.StatusResponse{
		Status: "initialized",
	}, nil
}

func (s *ChatServer) EditMessage(ctx context.Context, req *chatpb.EditMessageRequest) (*chatpb.SimpleResponse, error) {
	err := s.msgRepo.UpdateMessageByID(req.GetMessageId(), req.GetNewContent())
	if err != nil {
		return &chatpb.SimpleResponse{Message: "Failed to edit", Success: false}, err
	}
	return &chatpb.SimpleResponse{Message: "Message updated", Success: true}, nil
}

func (s *ChatServer) DeleteMessage(ctx context.Context, req *chatpb.DeleteMessageRequest) (*chatpb.SimpleResponse, error) {
	err := s.msgRepo.DeleteMessageByID(req.GetMessageId())
	if err != nil {
		return &chatpb.SimpleResponse{Message: "Failed to delete", Success: false}, err
	}
	return &chatpb.SimpleResponse{Message: "Message deleted", Success: true}, nil
}

func (s *ChatServer) SendTypingStatus(stream chatpb.ChatService_SendTypingStatusServer) error {
	for {
		status, err := stream.Recv()
		if err != nil {
			return err
		}
		if err := stream.Send(status); err != nil {
			return err
		}

		log.Printf("User %s typing to %s: %v", status.SenderId, status.ReceiverId, status.IsTyping)
	}
}

func (s *ChatServer) CreateGroup(ctx context.Context, req *chatpb.Group) (*chatpb.SimpleResponse, error) {
	err := s.msgRepo.CreateGroup(ctx, req.Id, req.GroupName, req.CreatorId, req.Members)
	if err != nil {
		return &chatpb.SimpleResponse{Message: "failed", Success: false}, err
	}
	return &chatpb.SimpleResponse{Message: "created", Success: true}, nil
}

func (s *ChatServer) AddGroupMember(ctx context.Context, req *chatpb.GroupMemberRequest) (*chatpb.SimpleResponse, error) {
	err := s.msgRepo.AddMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return &chatpb.SimpleResponse{Message: "failed", Success: false}, err
	}
	return &chatpb.SimpleResponse{Message: "added", Success: true}, nil
}

func (s *ChatServer) RemoveGroupMember(ctx context.Context, req *chatpb.GroupMemberRequest) (*chatpb.SimpleResponse, error) {
	err := s.msgRepo.RemoveGroupMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return &chatpb.SimpleResponse{Message: "failed", Success: false}, err
	}
	return &chatpb.SimpleResponse{Message: "removed", Success: true}, nil
}

func (s *ChatServer) GetGroupMembers(ctx context.Context, req *chatpb.GroupID) (*chatpb.GroupMembersResponse, error) {
	groupID := req.GroupId

	members, err := s.msgRepo.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &chatpb.GroupMembersResponse{Members: members}, nil
}

func (s *ChatServer) SendMessageRest(ctx context.Context, msg *chatpb.ChatMessage) (*chatpb.SendResponse, error) {
	log.Printf("REST message from %s to %s: %s", msg.SenderId, msg.ReceiverId, msg.Content)

	err := s.msgRepo.Save(ctx, msg)
	if err != nil {
		log.Printf("Save error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to save message: %v", err)
	}

	if conn, ok := chatws.Clients[msg.ReceiverId]; ok {
		if err := conn.Conn.WriteJSON(msg); err != nil {
			log.Printf("WebSocket send error: %v", err)
		}
	}

	return &chatpb.SendResponse{Status: "Message sent"}, nil
}
