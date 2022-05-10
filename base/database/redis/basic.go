package redis

import (
	"context"

	"github.com/go-redis/redis/extra/redisotel/v8"
	red "github.com/go-redis/redis/v8"
)

type Config struct {
	Addr     string `json:"addr"`
	Db       int    `json:"db"`
	Password string `json:"password"`
	Trace    bool   `json:"trace"`
}

type Basic interface {
	Set(ctx context.Context, key string, value string) (bool, error)
	Get(ctx context.Context, key string) (string, error)

	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

func NewClient(c Config) *red.Client {
	rdb := red.NewClient(&red.Options{
		Addr:     c.Addr,
		DB:       c.Db,
		Password: c.Password,
	})

	if c.Trace {
		rdb.AddHook(redisotel.NewTracingHook())
	}

	return rdb
}

type BasicRedis struct {
	*red.Client
}

func NewBasicRedis(r *red.Client) Basic {
	return &BasicRedis{Client: r}
}

func (r *BasicRedis) Set(ctx context.Context, key string, value string) (bool, error) {
	reply, err := r.Client.Set(ctx, key, value, 0).Result()

	return reply == "OK", err
}

func (r *BasicRedis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.Client.Get(ctx, key).Result()
	if err == red.Nil {
		return value, nil
	}
	return value, err
}

func (r *BasicRedis) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	v, err := r.Client.HSet(ctx, key, values...).Result()

	return v, err
}

func (r *BasicRedis) HGet(ctx context.Context, key, field string) (string, error) {
	v, err := r.Client.HGet(ctx, key, field).Result()
	if err == red.Nil {
		return v, nil
	}
	return v, err
}

func (r *BasicRedis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	v, err := r.Client.HGetAll(ctx, key).Result()
	if err == red.Nil {
		return map[string]string{}, err
	}
	return v, err
}
