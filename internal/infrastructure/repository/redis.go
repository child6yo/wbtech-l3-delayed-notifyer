package repository

import (
	"context"
	"time"

	z "github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/redis"
)

// Redis определяет подключение к Redis.
// Позволяет манипулировать данными в БД.
type Redis struct {
	client redis.Client
}

// NewRedis создает новый Redis.
func NewRedis(addr string, password string, db int) *Redis {
	return &Redis{client: *redis.New(addr, password, db)}
}

// Add добавляет новое значение по ключу.
func (r *Redis) Add(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return r.client.SetWithExpiration(ctx, key, value, exp)
}

// SortedSetAdd добавляет новое значение в SortedSet.
func (r *Redis) SortedSetAdd(ctx context.Context, set string, value interface{}, score float64) error {
	_, err := r.client.ZAdd(ctx, set, &z.Z{
		Score:  score,
		Member: value,
	}).Result()

	return err
}

// SortedSetRangeByScore возвращает диапазон значений из sorted set.
func (r *Redis) SortedSetRangeByScore(ctx context.Context, key, min, max string, offset, count int64) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &z.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

// Get возвращает значение по ключу.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

// Remove удаляет значение по ключу.
func (r *Redis) Remove(ctx context.Context, key string) error {
	return r.client.Del(ctx, key)
}

// SortedSetRemove удаляет значение из sorted set.
func (r *Redis) SortedSetRemove(ctx context.Context, set string, value interface{}) error {
	_, err := r.client.ZRem(ctx, set, value).Result()
	return err
}
