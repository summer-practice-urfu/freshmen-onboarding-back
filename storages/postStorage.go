package storages

import (
	"TaskService/db"
	"TaskService/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"log"
)

type PostStorage struct {
	conn         *pgx.Conn
	logger       *log.Logger
	tableName    string
	es           *db.EsDb
	esIndex      string
	esPostFields []string
}

func NewPostStorage(logger *log.Logger, conn *pgx.Conn, es *db.EsDb) *PostStorage {
	stor := &PostStorage{
		conn:         conn,
		logger:       logger,
		tableName:    "public.\"Posts\"",
		es:           es,
		esIndex:      "post",
		esPostFields: []string{"title", "content"},
	}
	stor.createTableIfNotExist()
	stor.createIndexIfNotExist()

	return stor
}

func (s *PostStorage) createIndexIfNotExist() {
	if err := s.es.CreateIndex(s.esIndex); err != nil {
		panic(err)
	}
	s.logger.Println("Created es index ", s.esIndex)
}

func (s *PostStorage) createTableIfNotExist() {
	_, err := s.conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS public.\"Posts\"\n"+
		"(\n   id uuid NOT NULL,"+
		"\n    title text COLLATE pg_catalog.\"default\" NOT NULL,"+
		"\n    content text COLLATE pg_catalog.\"default\","+
		"\n    rating integer NOT NULL DEFAULT 0,"+
		"\n    img text COLLATE pg_catalog.\"default\","+
		"\n    CONSTRAINT \"Posts_pkey\" PRIMARY KEY (id)"+
		"\n    )")
	if err != nil {

		panic(err)
	}
}

func (s *PostStorage) GetOne(id string) (*models.Post, error) {
	row, err := s.conn.Query(context.Background(), "select * from "+s.tableName+" where id=$1 limit 1;", id)
	if err != nil {
		return nil, err
	}
	var post models.Post
	if err := pgxscan.ScanOne(&post, row); err != nil {
		if pgxscan.NotFound(err) {
			return nil, err
		}
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
	var posts []models.Post
	if err := pgxscan.ScanAll(&posts, rows); err != nil {
		panic(err)
	}
	return posts, nil
}

func (s *PostStorage) GetMany(ids []string) ([]models.Post, error) {
	rows, err := s.conn.Query(context.Background(), "select * from "+s.tableName+" where id = ANY($1::uuid[])", ids)
	if err != nil {
		return nil, err
	}
	var tasks []models.Post
	if err := pgxscan.ScanAll(&tasks, rows); err != nil {
		panic(err)
	}
	if len(tasks) == 0 {
		tasks = make([]models.Post, 0)
	}

	return tasks, nil
}

func (s *PostStorage) Create(title, content, img string) (string, error) {
	id := uuid.New().String()
	_, err := s.conn.Exec(context.Background(), "Insert into "+s.tableName+
		" (id, title, content, img) values ($1, $2, $3, $4)", id, title, content, img)
	if err != nil {
		return "", nil
	}
	post := models.PostES{
		Id:      id,
		Title:   title,
		Content: content,
	}
	if err := s.es.Index(s.esIndex, id, post); err != nil {
		return "", nil
	}
	return id, err
}

func (s *PostStorage) ChangeRatingRelatively(id string, delta int) (int, error) {
	rows, err := s.conn.Query(context.Background(), "UPDATE "+s.tableName+
		"\n set rating=rating+$2"+
		"\n where id=$1"+
		"\n returning rating", id, delta)

	if err != nil {
		return 0, err
	}
	var newRating int
	if err := pgxscan.ScanOne(&newRating, rows); err != nil {
		return 0, err
	}

	return newRating, nil
}

func (s *PostStorage) Update(newPost *models.Post) error {
	_, err := s.conn.Exec(context.Background(), "update "+s.tableName+
		" set title=$2,"+
		" content=$3,"+
		" rating=$4,"+
		" img=$5 "+
		"where id=$1", newPost.Id, newPost.Title, newPost.Content, newPost.Rating, newPost.Img)

	if err != nil {
		return err
	}

	postES := &models.PostES{
		Id:    newPost.Id,
		Title: newPost.Title,
	}
	if newPost.Content != nil {
		postES.Content = *newPost.Content
	}

	err = s.es.Update(s.esIndex, newPost.Id, newPost)
	return err
}

func (s *PostStorage) Delete(id string) error {
	err := s.es.Delete(s.esIndex, id)
	if err != nil {
		return err
	}
	_, err = s.conn.Exec(context.Background(), "delete from "+s.tableName+
		"where id=$1", id)
	return err
}

func (s *PostStorage) SearchES(query string, size, page int) (*struct {
	Total int
	Ids   []string
}, error) {
	res, err := s.es.Search(s.esIndex, query, s.esPostFields, size, page)
	if err != nil {
		s.logger.Println("Error in SearchES, \nError: ", err.Error())
		return nil, err
	}

	ids := make([]string, len(res.Hits.Hits))
	for i, hit := range res.Hits.Hits {
		ids[i] = *hit.Id
	}

	return &struct {
		Total int
		Ids   []string
	}{Total: len(ids), Ids: ids}, nil
}
