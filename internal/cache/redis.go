package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr string, ttl time.Duration) (*Redis, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client, ttl: ttl}, nil
}

func (r *Redis) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set stores value with the configured cache TTL.
func (r *Redis) Set(ctx context.Context, key string, value any) error {
	return r.set(ctx, key, value, r.ttl)
}

// SetPermanent stores value with no expiry (used for durable store entries).
func (r *Redis) SetPermanent(ctx context.Context, key string, value any) error {
	return r.set(ctx, key, value, 0)
}

func (r *Redis) set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Scan returns all keys matching a pattern without blocking (safe for production).
func (r *Redis) Scan(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

func (r *Redis) Client() *redis.Client {
	return r.client
}
