package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Host       string
	Port       string
	Password   string
	DB         int
	DefaultTTL time.Duration
}

type Service struct {
	client     *redis.Client
	defaultTTL time.Duration
	enabled    bool
}

func NewService(ctx context.Context, opts Options) *Service {
	service := &Service{defaultTTL: opts.DefaultTTL}
	if opts.DefaultTTL <= 0 {
		service.defaultTTL = 5 * time.Minute
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       opts.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Printf("Redis is unavailable, cache disabled: %v", err)
		_ = client.Close()
		return service
	}

	service.client = client
	service.enabled = true
	return service
}

func (s *Service) Close() error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.client.Close()
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.client != nil
}

func (s *Service) Get(ctx context.Context, key string, dest any) (bool, error) {
	if !s.Enabled() {
		return false, nil
	}

	raw, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !s.Enabled() {
		return nil
	}

	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, raw, ttl).Err()
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	if !s.Enabled() || len(keys) == 0 {
		return nil
	}

	return s.client.Del(ctx, keys...).Err()
}

func (s *Service) DelByPattern(ctx context.Context, pattern string) error {
	if !s.Enabled() {
		return nil
	}

	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.client.Unlink(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			return nil
		}
	}
}
