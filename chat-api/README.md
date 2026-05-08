# Chat API

gRPC микросервис для peer-to-peer и группового обмена сообщениями. Поддерживает хранение истории в MongoDB, real-time доставку, а также REST-доступ через grpc-gateway. Реализован на Go с использованием Clean Architecture.

## Архитектура

```
chat-api/
├── cmd/
│   ├── static/                    # Загружаемые файлы (вложения)
│   └── main.go                    # Точка входа сервера
├── config/                        # Загрузка переменных окружения
│   └── config.go
├── db/
│   └── migrations/
│       ├── init_mongo.go         # Mongo миграции (создание индексов, seed)
│       └── cmd/main.go           # Отдельный запуск миграции
├── internal/
│   ├── delivery/
│   │   ├── grpc/server.go        # gRPC сервер
│   │   └── websocket/ws.go       # (если используется) WebSocket
│   └── repository/
│       └── message_repository.go # Работа с MongoDB
├── proto/
│   ├── chat.proto                # gRPC определения
│   └── gen/                      # Сгенерированные `.pb.go` файлы
├── .env                          # Переменные окружения
├── Dockerfile                    # Docker-сборка сервиса
├── docker-compose.yml           # Компоновка Mongo + chat-api
├── chat_client.go                # Тестовый клиент (опционально)
├── go.mod / go.sum               # Зависимости
```

## Быстрый старт

### Предварительные требования

- Go 1.23+
- MongoDB 5.0+
- [protoc](https://grpc.io/docs/protoc-installation/)
- `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`

### 1. Клонирование и установка

```bash
git clone <repository-url>
cd chat-api
go mod download
```

### 2. Настройка .env

```env
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=chatDB
```

### 3. Генерация `.pb.go` файлов

```bash
# Установка protoc (если не установлен)
# Windows: https://grpc.io/docs/protoc-installation/
# macOS: brew install protobuf
# Linux: apt-get install protobuf-compiler

# Установка плагинов (один раз)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Генерация кода
.\generate-proto.ps1
```

### 4. Применение миграций MongoDB

```bash
go run ./db/migrations/cmd/main.go
```

### 5. Запуск сервера

```bash
go run ./cmd/main.go
```

## Docker

### Сборка образа

```bash
docker build -t chat-api .
```

### Запуск контейнера

```bash
docker run -d -p 50051:50051 --env-file .env chat-api
```

## Docker Compose

```yaml
version: "3.9"

services:
  chat-api:
    build: .
    container_name: chat-api
    ports:
      - "50051:50051"
    env_file:
      - .env
    depends_on:
      - mongo

  mongo:
    image: mongo:6.0
    container_name: mongo
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

```bash
docker-compose up --build
```

## API Документация

### ChatService

| Метод            | Назначение                                  | Request                   | Response                  |
|------------------|----------------------------------------------|---------------------------|---------------------------|
| `SendMessage`    | Отправить peer-to-peer или групповое сообщение | `SendMessageRequest`      | `SendMessageResponse`     |
| `GetHistory`     | Получить историю сообщений по chat_id        | `GetHistoryRequest`       | `GetHistoryResponse`      |
| `CreateGroup`    | Создать новую группу                         | `CreateGroupRequest`      | `CreateGroupResponse`     |
| `ListGroups`     | Получить список групп                        | `ListGroupsRequest`       | `ListGroupsResponse`      |
| `GetGroupById`   | Получить данные по конкретной группе         | `GetGroupByIdRequest`     | `GetGroupByIdResponse`    |

### Примеры

#### 📩 SendMessage

```json
{
  "chat_id": "demo_group_8",
  "sender_id": "u3",
  "receiver_id": "u101",
  "content": "hello",
  "file_url": "/static/uploads/hello.png",
  "file_name": "hello.png",
  "file_type": "image"
}
```

#### 📜 GetHistory

```json
{
  "chat_id": "demo_group_8",
  "limit": 50,
  "offset": 0
}
```

#### 👥 CreateGroup

```json
{
  "group_id": "demo_group_8",
  "group_name": "Demo",
  "creator_id": "u1",
  "members": ["u3", "u5", "u7", "u101"]
}
```

## Переменные окружения

| Переменная      | Назначение                       |
|------------------|----------------------------------|
| `MONGO_URI`      | URI подключения к MongoDB        |
| `MONGO_DATABASE` | Название базы данных             |