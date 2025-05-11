package redis

import (
	"context"
	"fmt"
	"github.com/kxddry/url-shortener/internal/config"
	"github.com/kxddry/url-shortener/internal/lib/random"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisOptions struct {
	redis.Options
}

type RedisClient struct {
	redis.Options
	client *redis.Client
}

func NewRedisClient(cfg config.RedisStorage) *RedisClient {
	return &RedisClient{
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
}

func (r *RedisClient) Connect() error {
	if r.client != nil {
		return nil
	}
	r.client = redis.NewClient(&r.Options)
	return r.client.Ping(context.Background()).Err()
}

func (r *RedisClient) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.redis.SaveURL"
	multi := r.client.TxPipeline()

	err := multi.SetNX(context.Background(), alias, urlToSave, 0).Err()
	if err != nil {
		multi.Discard()
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	multi.Save(context.Background())
	return 1337, nil // return 1337 to indicate redis is used
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

func (r *RedisClient) GenerateAlias(i int) (string, error) {
	const op = "storage.redis.GenerateAlias"

	alias := random.NewRandomString(i)
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		if ctx.Err() != nil {
			return "", fmt.Errorf("%s: %w", op, ctx.Err())
		}
		err = r.client.SetNX(ctx, alias, "", 0).Err()
		if err == nil {
			break
		}
		alias = random.NewRandomString(i)
	}
	return alias, nil
}
