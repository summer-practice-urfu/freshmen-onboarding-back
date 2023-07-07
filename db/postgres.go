package db

import (
	"context"
	pgx "github.com/jackc/pgx/v5"
	"log"
	"os"
	"time"
)

type PostgresDb struct {
	Conn   *pgx.Conn
	logger *log.Logger
}

func Init(logger *log.Logger) *PostgresDb {
	url := os.Getenv("DATABASE_URL")
	logger.Println("url", url)
	conn, err := pgx.Connect(context.Background(), url)
	countRetry := 5
	for err != nil && countRetry > 0 {
		time.Sleep(time.Second * 1)
		conn, err = pgx.Connect(context.Background(), url)
		countRetry--
	}
	if err != nil {
		panic(err)
	}
	logger.Println("Connected to postgres")
	return &PostgresDb{Conn: conn, logger: logger}
}

func (d *PostgresDb) Close() {
	if err := d.Conn.Close(context.Background()); err != nil {
		panic(err)
	}
}
