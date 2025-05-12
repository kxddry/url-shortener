package redis

import (
	"context"
	"fmt"
	"github.com/kxddry/url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisOptions struct {
	redis.Options
}

type RedisClient struct {
	redis.Options
	client *redis.Client
}

func New(cfg config.RedisStorage) (*RedisClient, error) {
	r := RedisClient{
		Options: redis.Options{
			Addr:     cfg.Host + ":" + cfg.Port,
			Password: cfg.Password,
			Username: cfg.User,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
			Network:  cfg.Protocol,
		},
		client: nil,
	}
	return &r, r.connect()
}

func (r *RedisClient) connect() error {
	if r.client != nil {
		return nil
	}
	r.client = redis.NewClient(&r.Options)
	return r.client.Ping(context.Background()).Err()
}

func (r *RedisClient) SaveURL(urlToSave, alias string, creator int64) (int64, error) {
	_ = creator // don't store the creator in redis
	const op = "storage.redis.SaveURL"
	multi := r.client.TxPipeline()

	err := multi.SetNX(context.Background(), alias, urlToSave, 0).Err()
	if err != nil {
		multi.Discard()
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	multi.Save(context.Background())
	return 0, nil
}

func (r *RedisClient) GetURL(alias string) (string, error) {
	const op = "storage.redis.GetURL"
	url, err := r.client.Get(context.Background(), alias).Result()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (r *RedisClient) DeleteURL(alias string) error {
	const op = "storage.redis.DeleteURL"
	err := r.client.Del(context.Background(), alias).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
