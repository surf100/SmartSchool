# API Gateway

API Gateway для микросервисной архитектуры, построенный на базе gRPC-Gateway.

## Установка зависимостей

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

## Инициализация проекта

```bash
go mod init api-gateway
go mod tidy
```

## Подготовка Google APIs

```bash
git clone https://github.com/googleapis/googleapis.git
```

## Генерация proto файлов

```bash
protoc -I proto \
       -I ./googleapis \
       --go_out=proto/gen --go_opt=paths=source_relative \
       --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=proto/gen --grpc-gateway_opt=paths=source_relative \
       proto/*.proto
```

## Запуск

```bash
go run cmd/main.go
```

API Gateway будет доступен на порту 8081.

## Подключенные сервисы

- **Person Service** (localhost:50053) - Управление данными о персонах
- **Canteen Service** (localhost:50052) - Управление питанием и платежами
- **Library Service** (localhost:50054) - Управление библиотекой

## API Endpoints

### Person API

- `POST /api/persons` - Создать нового человека
- `GET /api/persons/{id}` - Получить человека по ID
- `GET /api/persons` - Получить список всех людей
- `PUT /api/persons` - Обновить данные человека
- `DELETE /api/persons/{id}` - Удалить человека по ID

### Изменения в Person API

- ID остается автоинкрементным первичным ключом (1, 2, 3...)
- PIN добавлен как уникальное поле для дополнительной идентификации
- Обновлена схема данных согласно новым требованиям
- Добавлены новые поля: inSchool, susn, parents

## Примеры запросов

### Создание человека
```bash
curl -X POST http://localhost:8081/api/persons \
  -H "Content-Type: application/json" \
  -d '{
    "pin": "123456",
    "name": "Иван",
    "lastName": "Иванов",
    "email": "ivanov@example.com",
    "isDisabled": false
  }'
```

### Получение человека по ID
```bash
curl http://localhost:8081/api/persons/1
```

## Структура проекта

```
api-gateway/
├── cmd/main.go              # Точка входа
├── proto/                   # Proto файлы
│   ├── person.proto         # Person API
│   ├── canteen.proto        # Canteen API
│   ├── library.proto        # Library API
│   └── gen/                 # Сгенерированные файлы
└── README.md               # Документация
```

