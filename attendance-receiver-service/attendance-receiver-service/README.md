# Attendance API

Микросервис для обработки событий от системы ZKBio CVSecurity и сохранения данных о посещаемости.

## Описание

Attendance API принимает HTTP POST-запросы от системы ZKBio и сохраняет данные о входах пользователей в таблицу `attendancelog`. Сервис работает через gRPC и интегрируется с API Gateway.

## Архитектура

Сервис построен по принципу Clean Architecture с разделением на слои:

- **Domain** - доменные модели и интерфейсы
- **Application** - интерфейсы сервисов
- **UseCase** - бизнес-логика
- **Infrastructure** - реализация репозиториев
- **Adapters** - gRPC сервер

## Установка и запуск

### Предварительные требования

- Go 1.24+
- PostgreSQL
- protoc (Protocol Buffers compiler)

### Настройка базы данных

1. Создайте базу данных PostgreSQL
2. Выполните миграции из папки `db/migrations/`

```sql
-- Создание таблицы attendancelog
CREATE TABLE IF NOT EXISTS attendancelog (
    id SERIAL PRIMARY KEY,
    pin TEXT,
    name TEXT,
    last_name TEXT,
    dept_code TEXT,
    dept_name TEXT,
    reader_name TEXT,
    door_name TEXT,
    device_sn TEXT,
    capture_photo_path TEXT,
    verify_mode_name TEXT,
    event_name TEXT,
    event_time TIMESTAMP,
    raw_payload JSONB,
    unique_key TEXT UNIQUE,
    log_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Настройка окружения

1. Скопируйте `.example.env` в `.env`
2. Настройте переменные окружения:

```env
POSTGRES_DSN=host=localhost port=5432 user=postgres password=password dbname=attendance sslmode=disable
```

### Генерация proto файлов

```powershell
./scripts/gen.ps1
```

### Запуск сервиса

```bash
go run cmd/main.go
```

Сервис запустится на порту `:50055`.

## API Endpoints

### gRPC методы

- `PushEvent` - обработка событий от ZKBio
- `GetAttendanceLogs` - получение логов посещаемости

### HTTP endpoints (через API Gateway)

- `POST /api/attendance/zkbio` - прием событий от ZKBio
- `GET /api/attendance/logs` - получение логов посещаемости

## Формат данных ZKBio

ZKBio отправляет HTTP POST-запрос с JSON-объектом:

```json
{
  "module": "acc",
  "dataType": "text",
  "moduleName": "Access",
  "content": "{...}", // Вложенный JSON как строка
  "pushType": "7",
  "pushTypeName": "Event Log"
}
```

Поле `content` содержит JSON с деталями события:

```json
{
  "areaId": "402888479817e873019817ed2ca30003",
  "areaName": "Area Name",
  "capturePhotoPath": "/upload/event/photo/QJT3244400420/20250730232934.jpg",
  "cardNo": "",
  "deptCode": "2",
  "deptId": "40288847984067610198406f9eed0018",
  "deptName": "11A",
  "description": "",
  "devAlias": "192.168.8.198",
  "devId": "402888479817e873019817efd1cf0bc3",
  "devSn": "QJT3244400420",
  "doorId": "402888479817e873019817efd2800c12",
  "doorName": "192.168.8.198-1",
  "equals": false,
  "eventAddr": 1,
  "eventLevel": 0,
  "eventName": "acc_newEventNo_1",
  "eventNo": 1,
  "eventPointId": "402888479817e873019817efd2800c12",
  "eventPointName": "192.168.8.198-1",
  "eventPointType": 0,
  "eventTime": 1753943374000,
  "lastName": "Синченко",
  "logId": 95,
  "maskFlag": "0",
  "name": "Александр",
  "persPersonId": "402888479817e8730198181126530d03",
  "pin": "123",
  "readerId": "402888479817e873019817efd2820c14",
  "readerName": "192.168.8.198-1-Out",
  "readerState": 1,
  "temperature": "0",
  "uniqueKey": "QJT3244400420_95_2025-07-30 23:29:34",
  "verifyModeName": "acc_verify_mode_onlyface",
  "verifyModeNo": 15
}
```

## Структура базы данных

Таблица `attendancelog` содержит следующие поля:

- `id` - уникальный идентификатор записи
- `pin` - PIN-код пользователя
- `name` - имя пользователя
- `last_name` - фамилия пользователя
- `dept_code` - код отдела
- `dept_name` - название отдела
- `reader_name` - название считывателя
- `door_name` - название двери
- `device_sn` - серийный номер устройства
- `capture_photo_path` - путь к фотографии
- `verify_mode_name` - название режима верификации
- `event_name` - название события
- `event_time` - время события
- `raw_payload` - оригинальный JSON события
- `unique_key` - уникальный ключ события
- `log_id` - ID лога от ZKBio
- `created_at` - время создания записи

## Защита от дубликатов

Сервис проверяет уникальность событий по полю `unique_key` и пропускает дубликаты.

## Интеграция с ZKBio

Для настройки интеграции с ZKBio:

1. Перейдите в интерфейсе ZKBio: Service Center → Push Center → Push Configuration → Add New
2. Настройте параметры:
   - Name: Attendance Push
   - URL: http://{GRPC_API_GATEWAY}/api/attendance/zkbio
   - Checkboxes: ✅ Event Log, ✅ Capture Photo, ✅ Attendance Record, ✅ Attendance Photo

## Разработка

### Структура проекта

```
attendance-api/
├── cmd/
│   └── main.go                 # Точка входа
├── internal/
│   ├── adapters/
│   │   └── grpc/
│   │       └── server.go       # gRPC сервер
│   ├── domain/
│   │   ├── application/
│   │   │   └── attendance_service.go
│   │   ├── repository/
│   │   │   └── attendance_repository.go
│   │   └── attendance.go       # Доменные модели
│   ├── infrastructure/
│   │   └── postgres/
│   │       └── attendance_postgres.go
│   └── usecase/
│       └── attendance_usecase.go
├── proto/
│   └── attendance.proto        # Protocol Buffers определение
├── db/
│   └── migrations/             # Миграции базы данных
└── scripts/
    └── gen.ps1                 # Скрипт генерации proto
``` 