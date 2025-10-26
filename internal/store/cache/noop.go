package cache

import (
	"context"
	"tiago-udemy/internal/store"
)

// noOpUserCache implements User interface but does nothing
type noOpUserCache struct{}

func (n *noOpUserCache) Set(ctx context.Context, user *store.User) error {
	return nil // Always succeed but do nothing
}

func (n *noOpUserCache) Get(ctx context.Context, userId int64) (*store.User, bool, error) {
	return nil, false, nil // Always return cache miss
}

func (n *noOpUserCache) Delete(ctx context.Context, userID int64) error {
	return nil // Always succeed but do nothing
}

// NewNoOpStore returns a CacheStorage that does nothing
func NewNoOpStore() CacheStorage {
	return CacheStorage{
		Users: &noOpUserCache{},
	}
}
