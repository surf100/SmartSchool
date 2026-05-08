package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"chat-api/internal/repository"
	"chat-api/proto/gen"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var MsgRepo *repository.MessageRepository

type Client struct {
	Conn   *websocket.Conn
	userID string
}

var Clients = make(map[string]*Client)
var mu sync.Mutex

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	client := &Client{Conn: conn, userID: userID}
	mu.Lock()
	Clients[userID] = client
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(Clients, userID)
		mu.Unlock()
		conn.Close()
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var msg gen.ChatMessage
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}
		log.Printf("Saving message: sender=%s receiver=%s content=%s isGroup=%t groupID=%s",
			msg.SenderId, msg.ReceiverId, msg.Content, msg.IsGroup, msg.ChatId)

		err = MsgRepo.Save(r.Context(), &msg)

		if err != nil {
			log.Println("Failed to save message:", err)
		}
		if msg.Content == "" {
			log.Println("Content is empty, skipping insert.")
		}

		if msg.IsGroup {
			groupMembers, err := MsgRepo.GetGroupMembers(r.Context(), msg.ChatId)
			if err != nil {
				log.Println("GetGroupMembers error:", err)
				continue
			}
			for _, memberID := range groupMembers {
				if memberID == msg.SenderId {
					continue
				}
				mu.Lock()
				if client, ok := Clients[memberID]; ok {
					client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
				}
				mu.Unlock()
			}
		} else {
			mu.Lock()
			if receiverClient, ok := Clients[msg.ReceiverId]; ok {
				receiverClient.Conn.WriteMessage(websocket.TextMessage, msgBytes)
			}
			mu.Unlock()
		}
	}
}

func StartWebSocketServer(repo *repository.MessageRepository) {
	MsgRepo = repo
	http.HandleFunc("/ws", HandleWebSocket)
	log.Println("WebSocket server started at :8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
