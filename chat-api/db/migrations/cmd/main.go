package main

import (
	"log"

	"chat-api/db/migrations"
)

func main() {
	mongoURI := "mongodb://localhost:27017"
	if err := migrations.RunMongoMigrations(mongoURI); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
