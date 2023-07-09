package storages

import (
	"TaskService/db"
	"TaskService/models"
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"log"
)

var (
	OperPlus  = '+'
	OperMinus = '-'
	OperNone  = '0'
	opers     = []rune{OperPlus, OperMinus, OperNone}
)

type UserPostRatingStorage struct {
	conn      *pgx.Conn
	logger    *log.Logger
	tableName string
}

func NewUserPostRatingStorage(logger *log.Logger, conn *pgx.Conn, es *db.EsDb) *UserPostRatingStorage {
	stor := &UserPostRatingStorage{
		conn:      conn,
		logger:    logger,
		tableName: "public.\"UserPostRating\"",
	}
	stor.createTableIfNotExist()

	return stor
}

func (s *UserPostRatingStorage) createTableIfNotExist() {
	_, err := s.conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS "+s.tableName+
		"\n ("+
		"\n \"userId\" integer NOT NULL,"+
		"\n \"postId\" uuid NOT NULL,"+
		"\n oper \"char\" NOT NULL,"+
		"\n CONSTRAINT \"UserPostRating_pkey\" PRIMARY KEY (\"userId\", \"postId\"),"+
		"\n CONSTRAINT \"UserPostRating_postId_fkey\" FOREIGN KEY (\"postId\")"+
		"\n REFERENCES public.\"Posts\" (id) MATCH SIMPLE"+
		"\n ON UPDATE NO ACTION"+
		"\n ON DELETE CASCADE"+
		"\n )")
	if err != nil {
		panic(err)
	}
}

func (s *UserPostRatingStorage) GetUserOper(userId int, postId string) (*models.UserPostRating, error) {
	row, err := s.conn.Query(context.Background(), "SELECT \"userId\", \"postId\", oper "+
		"\n FROM public.\"UserPostRating\""+
		"\n WHERE \"postId\" = $1"+
		"\n AND \"userId\" = $2;", postId, userId)
	if err != nil {
		return nil, err
	}
	var userPostRating models.UserPostRating
	if err := pgxscan.ScanOne(&userPostRating, row); err != nil {
		if pgxscan.NotFound(err) {
			return nil, err
		}
	}

	return &userPostRating, nil
}

func (s *UserPostRatingStorage) SetUserOper(userPostRating *models.UserPostRating) error {
	if !s.operAllowed(userPostRating.Oper) {
		return errors.New("invalid oper")
	}

	existing, err := s.GetUserOper(userPostRating.UserId, userPostRating.PostId)

	if pgxscan.NotFound(err) {
		return s.CreateUserOper(userPostRating)
	}

	existing.Oper = userPostRating.Oper
	return s.UpdateUserOper(userPostRating)
}

func (s *UserPostRatingStorage) UpdateUserOper(newUserPostRating *models.UserPostRating) error {
	if !s.operAllowed(newUserPostRating.Oper) {
		return errors.New("invalid oper")
	}

	_, err := s.conn.Exec(context.Background(), "UPDATE "+s.tableName+
		"\n SET oper=$3"+
		"\n WHERE \"userId\"=$1 AND \"postId\"=$2", newUserPostRating.UserId, newUserPostRating.PostId, newUserPostRating.Oper)

	return err
}

func (s *UserPostRatingStorage) CreateUserOper(userPostRating *models.UserPostRating) error {
	if !s.operAllowed(userPostRating.Oper) {
		return errors.New("invalid oper")
	}

	_, err := s.conn.Exec(context.Background(), "INSERT INTO "+s.tableName+" ("+
		"\n\t \"userId\", \"postId\", oper)"+
		"\n\t  VALUES ($1, $2, $3);", userPostRating.UserId, userPostRating.PostId, userPostRating.Oper)

	return err
}

func (s *UserPostRatingStorage) operAllowed(oper rune) bool {
	allowed := false
	for _, allowedOper := range opers {
		if oper == allowedOper {
			allowed = true
			break
		}
	}
	return allowed
}
