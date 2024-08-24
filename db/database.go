package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewDatabaseConnection() (*Database, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	connString := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
