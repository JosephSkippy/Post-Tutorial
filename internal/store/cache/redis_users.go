package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"tiago-udemy/internal/store"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisUserCache struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisUserCache(rdb *redis.Client, ttl time.Duration) *RedisUserCache {
	if rdb == nil {
		panic("redis client cannot be nil")
	}
	if ttl <= 0 {
		ttl = CacheDefaultTTL
	}
	return &RedisUserCache{rdb: rdb, prefix: "user:", ttl: ttl}
}

func (c *RedisUserCache) cacheKey(id int64) string {
	return fmt.Sprintf("%s%d", c.prefix, id)
}

func (c *RedisUserCache) Get(ctx context.Context, userId int64) (*store.User, bool, error) {

	val, err := c.rdb.Get(ctx, c.cacheKey(userId)).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil // cache miss
		}
		return nil, false, err // redis error
	}

	user := &store.User{}
	if err := json.Unmarshal(val, user); err != nil {
		_ = c.rdb.Del(ctx, c.cacheKey(userId)).Err()
		return nil, false, err
	}

	return user, true, nil
}

func (c *RedisUserCache) Set(ctx context.Context, user *store.User) error {

	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// TTL jitter (0–30s) to prevent herd expiry
	jitter := time.Duration(rand.Intn(30)) * time.Second
	if err := c.rdb.Set(ctx, c.cacheKey(user.ID), userJson, c.ttl+jitter).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisUserCache) Delete(ctx context.Context, userID int64) error {
	return c.rdb.Del(ctx, c.cacheKey(userID)).Err()
}
