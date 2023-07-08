package storages

import (
	"TaskService/db"
	"TaskService/models/cache"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

var (
	redisKeySessionToken = "session_storage"
)

type SessionStorage struct {
	redis  *db.RedisDb
	logger *log.Logger
}

func NewSessionStorage(redis *db.RedisDb, logger *log.Logger) *SessionStorage {
	stor := &SessionStorage{
		redis:  redis,
		logger: logger,
	}

	return stor
}

func (s *SessionStorage) CreateSession(userVk *cache.UserVk) (string, error) {
	sessionToken := uuid.New().String()
	sessionKey := makeSessionKey(sessionToken)
	duration := userVk.Token.Expiry.Sub(time.Now())
	s.logger.Println("Setting token: ", sessionToken, " with duration: ", duration)
	err := s.redis.Set(sessionKey, userVk, duration)
	return sessionToken, err
}

func (s *SessionStorage) DeleteSession(sessionToken string) error {
	sessionKey := makeSessionKey(sessionToken)
	return s.redis.Delete(sessionKey)
}

func (s *SessionStorage) GetSession(sessionToken string) (*cache.UserVk, error) {
	key := makeSessionKey(sessionToken)
	val, err := s.redis.Get(key)
	if err != nil {
		return nil, errors.New("key not found")
	}

	var userVk *cache.UserVk
	if err := json.NewDecoder(strings.NewReader(val)).Decode(&userVk); err != nil {
		return nil, errors.New("error decoding token")
	}
	return userVk, nil
}

func parseSessionKey(sessionKey string) (string, error) {
	items := strings.Split(sessionKey, db.RedisDelimeter)
	if len(items) > 1 {
		return items[1], nil
	}
	return "", errors.New("invalid session key")
}

func makeSessionKey(sessionToken string) string {
	return redisKeySessionToken + db.RedisDelimeter + sessionToken
}
