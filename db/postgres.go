package db

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
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
	baseConn, err := pgx.Connect(context.Background(), url)
	countRetry := 5
	for err != nil && countRetry > 0 {
		time.Sleep(time.Second * 1)
		baseConn, err = pgx.Connect(context.Background(), url)
		countRetry--
	}
	if err != nil {
		panic(err)
	}
	logger.Println("Connected to postgres")
	if !baseExists(baseConn) {
		createBase(baseConn)
	}
	conn, err := pgx.Connect(context.Background(), url+"/summerPractice")
	countRetry = 5
	for err != nil && countRetry > 0 {
		time.Sleep(time.Second * 1)
		baseConn, err = pgx.Connect(context.Background(), url+"/summerPractice")
		countRetry--
	}
	if err != nil {
		panic(err)
	}
	db := &PostgresDb{Conn: conn, logger: logger}
	return db
}

func (d *PostgresDb) Close() {
	if err := d.Conn.Close(context.Background()); err != nil {
		panic(err)
	}
}

func baseExists(conn *pgx.Conn) bool {
	rows, errQuery := conn.Query(context.Background(), "SELECT datname FROM pg_catalog.pg_database "+
		"WHERE lower(datname) = lower('summerPractice');")
	var res []string
	if err := pgxscan.ScanAll(&res, rows); err != nil || errQuery != nil {
		panic(err)
	}
	return len(res) > 0
}

func createBase(conn *pgx.Conn) {
	_, err := conn.Exec(context.Background(), "CREATE DATABASE \"summerPractice\"\n"+
		"    WITH\n    OWNER = postgres\n"+
		"    ENCODING = 'UTF8'\n"+
		"    CONNECTION LIMIT = -1\n"+
		"    IS_TEMPLATE = False;")
	if err != nil {
		panic(err)
	}
}
