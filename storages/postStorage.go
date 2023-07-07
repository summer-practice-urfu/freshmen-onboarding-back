package storages

import (
	"TaskService/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"log"
)

type PostStorage struct {
	conn      *pgx.Conn
	logger    *log.Logger
	tableName string
}

func NewPostStorage(conn *pgx.Conn, logger *log.Logger) *PostStorage {
	return &PostStorage{
		conn:      conn,
		logger:    logger,
		tableName: "public.\"Posts\"",
	}
}

func (s *PostStorage) GetOne(id string) (*models.Post, error) {
	row, err := s.conn.Query(context.Background(), "select * from "+s.tableName+" where id=$1 limit 1;", id)
	if err != nil {
		return nil, err
	}
	var post models.Post
	if err := pgxscan.ScanOne(&post, row); err != nil {
		panic(err)
	}
	return &post, nil
}

func (s *PostStorage) GetAll(limit int) ([]models.Post, error) {
	rows, err := s.conn.Query(context.Background(), "select * from "+s.tableName+" limit $1", limit)
	if err != nil {
		s.logger.Println("Error in GetAll() \nError: ", err.Error())
		return nil, err
	}
	var tasks []models.Post
	if err := pgxscan.ScanAll(&tasks, rows); err != nil {
		panic(err)
	}
	return tasks, nil
}

func (s *PostStorage) GetMany(ids []int) ([]models.Post, error) {
	rows, err := s.conn.Query(context.Background(), "select * from "+s.tableName+" where id = ANY($1::int[])", ids)
	if err != nil {
		return nil, err
	}
	var tasks []models.Post
	if err := pgxscan.ScanAll(&tasks, rows); err != nil {
		panic(err)
	}
	return tasks, nil
}

func (s *PostStorage) Create(title, content string, img *string) (string, error) {
	id := uuid.New().String()
	_, err := s.conn.Exec(context.Background(), "Insert into "+s.tableName+
		" (id, title, content, img) values ($1, $2, $3, $4)", id, title, content, img)
	return id, err
}

func (s *PostStorage) Update(newPost *models.Post) error {
	_, err := s.conn.Exec(context.Background(), "update "+s.tableName+
		" set title=$2,"+
		" content=$3,"+
		" img=$4 "+
		"where id=$1", newPost.Id, newPost.Title, newPost.Content, newPost.Img)
	return err
}

func (s *PostStorage) Delete(id string) error {
	_, err := s.conn.Exec(context.Background(), "delete from "+s.tableName+
		"where id=$1", id)
	return err
}
