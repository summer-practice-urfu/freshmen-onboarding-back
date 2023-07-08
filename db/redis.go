package db

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

var (
	RedisDelimeter = "::"
)

type RedisDb struct {
	client *redis.Client
}

func NewRedisDb() (*RedisDb, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := client.Ping(context.Background()).Err()

	return &RedisDb{
		client: client,
	}, err
}

func (r *RedisDb) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration).Err()
}

func (r *RedisDb) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *RedisDb) Delete(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}
