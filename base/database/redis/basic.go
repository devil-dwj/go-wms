package redis

import (
	"context"
	"time"

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
	Exists(ctx context.Context, keys ...string) (bool, error)
	ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error)

	Set(ctx context.Context, key string, value string, t time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) (bool, error)

	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, field string) error
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
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

func (r *BasicRedis) Exists(ctx context.Context, keys ...string) (bool, error) {
	i, err := r.Client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}

	return i == 1, nil
}

func (r *BasicRedis) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	return r.Client.ExpireAt(ctx, key, tm).Result()
}

func (r *BasicRedis) Set(ctx context.Context, key string, value string, t time.Duration) (bool, error) {
	reply, err := r.Client.Set(ctx, key, value, t).Result()

	return reply == "OK", err
}

func (r *BasicRedis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.Client.Get(ctx, key).Result()
	if err == red.Nil {
		return "", nil
	}
	return value, err
}

func (r *BasicRedis) Del(ctx context.Context, key ...string) (bool, error) {
	i, err := r.Client.Del(ctx, key...).Result()
	if err != nil {
		return false, err
	}

	return i == 1, nil
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

func (r *BasicRedis) HDel(ctx context.Context, key string, field string) error {
	err := r.Client.HDel(ctx, key, field).Err()
	if err == red.Nil {
		return nil
	}
	return err
}

func (r *BasicRedis) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	v, err := r.Client.HIncrBy(ctx, key, field, incr).Result()
	return v, err
}
