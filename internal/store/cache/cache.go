package cache

import (
	"context"
	"tiago-udemy/internal/store"
	"time"

	"github.com/go-redis/redis/v8"
)

const CacheDefaultTTL = 10 * time.Minute

type User interface {
	Set(ctx context.Context, user *store.User) error
	Get(ctx context.Context, userId int64) (*store.User, bool, error)
	Delete(ctx context.Context, userID int64) error
}

type CacheStorage struct {
	Users User
}

func RedisStore(rdb *redis.Client) CacheStorage {
	return CacheStorage{
		Users: NewRedisUserCache(rdb, CacheDefaultTTL),
	}
}
