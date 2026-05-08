# SmartSchool Microservices

Backend microservices platform for educational institutions.  
The system includes user synchronization, biometric access control integration, school meal voucher services, internal chat, and API gateway routing.

## Overview

SmartSchool is a microservices-based backend system developed for educational organizations.  
It connects several school-related services into one backend platform and provides REST and gRPC APIs for frontend applications and external systems.

The project includes integration with ZKBio Security for biometric data synchronization, PostgreSQL-based user and school data management, MongoDB-based chat storage, and RabbitMQ for asynchronous communication between services.

Main areas covered by the system:

- user and person management;
- biometric user synchronization;
- peer-to-peer and group chat;
- school meal voucher activation and transaction history;
- API gateway for unified REST access;
- service-to-service communication through gRPC.

---

## My Role

As a backend developer, I worked on the design and implementation of the backend services.

Main responsibilities:

- designed the microservice structure and service boundaries;
- implemented backend services in Go;
- configured gRPC communication between services;
- exposed REST endpoints through gRPC-Gateway;
- implemented integration with ZKBio Security;
- worked on the Person Dumper service for user synchronization;
- designed PostgreSQL database structures, including multi-schema logic;
- used MongoDB for chat message storage;
- configured RabbitMQ for asynchronous communication;
- prepared Swagger documentation for API endpoints;
- containerized services using Docker and Docker Compose.

---

## Screenshots

### Dashboard
![Dashboard](assets/Main.png)

---

## Architecture

```text
+------------------+
|   Frontend UI    |
+--------+---------+
         |
         | REST API
         v
+------------------+
|   API Gateway    |
|  gRPC-Gateway    |
+--+---+---+---+---+
   |   |   |   |
   v   v   v   v
Person Chat Social Person
 API   API Wallet Dumper
  |     |     |      |
  v     v     v      v
PostgreSQL MongoDB PostgreSQL ZKBio Security
        |
        v
     RabbitMQ
```

The API Gateway is used as the main entry point for frontend clients.  
Internal services communicate through gRPC. RabbitMQ is used for asynchronous events.  
The Person Dumper service synchronizes data from ZKBio Security and sends the processed user data to the Person API.

---

## Services

| Service | Responsibility | Storage |
|---|---|---|
| API Gateway | REST entry point, request routing, health checks | - |
| Person API | User management, students, teachers, admins, SCUD-related logic | PostgreSQL |
| Person Dumper | Synchronization from ZKBio Security to Person API | PostgreSQL |
| Chat API | Peer-to-peer and group messaging, offline messages, history | MongoDB |
| Social Wallet Service | Meal voucher activation, deactivation, transaction history | PostgreSQL |

---

## Features

### User Synchronization

- Synchronization with ZKBio Security
- Import of users from biometric/access control systems
- Department and school relationship mapping
- Multi-schema PostgreSQL structure for school-level separation
- Support for background synchronization through Person Dumper

### Chat Service

- Peer-to-peer messaging
- Group messaging
- Offline message storage
- Message delivery after reconnect
- Message history
- Search and filtering support
- gRPC and REST access through gateway

### Social Wallet Service

- Student meal voucher activation
- Voucher deactivation
- Transaction history retrieval
- Integration with external social wallet services
- PostgreSQL-based persistence

### API Gateway

- Single REST entry point for frontend clients
- Routing requests to internal gRPC services
- Gateway proto files with HTTP annotations
- Health check endpoints
- Swagger documentation support

### Person Management

- Student, teacher, and admin data handling
- SCUD-related authentication flow support
- User data synchronization from external systems
- PostgreSQL-based school data storage
- Multi-schema organization for separating school data

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| Service Communication | gRPC |
| REST Access | gRPC-Gateway |
| API Documentation | Swagger |
| Main Database | PostgreSQL |
| Chat Storage | MongoDB |
| Message Broker | RabbitMQ |
| Containerization | Docker, Docker Compose |
| External Integration | ZKBio Security API |

---

## Getting Started

### Requirements

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 14+
- MongoDB 6+
- RabbitMQ 3+
- ZKBio Security instance for synchronization features

### Run with Docker Compose

```bash
git clone https://github.com/your-username/smartschool-microservices.git
cd smartschool-microservices
cp .env.example .env
docker-compose up --build
```

### Run a Service Locally

```bash
cd services/person-api
go run cmd/main.go
```

---

## Environment Variables

Create a `.env` file based on `.env.example`.

Example variables:

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=smartschool

MONGO_URI=mongodb://localhost:27017
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

PERSON_API_GRPC_ADDR=localhost:50053
API_GATEWAY_PORT=8080

ZKBIO_BASE_URL=http://localhost:8098
ZKBIO_TOKEN=your_token
```

Do not upload `.env` files or real credentials to public repositories.

---

## Project Structure

```text
smartschool/
├── api-gateway/
├── person-api/
├── person-dumper/
├── chat-api/
├── social-wallet-service/
├── proto/
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Security Notes

- Real API tokens and passwords must be stored only in environment variables.
- `.env` files should not be committed.
- External service credentials must be replaced with placeholders before uploading to GitHub.
- Swagger documentation should not expose internal secrets or private endpoints.
- Production deployments should use separate credentials for each service.

---

## Future Improvements

- Prometheus metrics for each service
- Grafana dashboards
- Centralized structured logging
- Kubernetes deployment configuration
- Authentication service with JWT and refresh tokens
- Library management service
- Push notification service
- Admin panel API for school management
