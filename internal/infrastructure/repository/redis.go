package repository

import (
	"context"

	z "github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/redis"
)

type Redis struct {
	client redis.Client
}

func NewRedis(client redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) Add(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value)
}

func (r *Redis) SortedSetAdd(ctx context.Context, set string, value interface{}, score float64) error {
	_, err := r.client.ZAdd(ctx, set, &z.Z{
		Score:  score,
		Member: value,
	}).Result()

	return err
}

func (r *Redis) SortedSetRangeByScore(ctx context.Context, key, min, max string, offset, count int64) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &z.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

func (r *Redis) Remove(ctx context.Context, key string) error {
	return r.client.Del(ctx, key)
}

func (r *Redis) SortedSetRemove(ctx context.Context, set string, value interface{}) error {
	_, err := r.client.ZRem(ctx, set, value).Result()
	return err
}
