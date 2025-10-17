package poller

import (
	"context"
	"strconv"
	"time"
)

type storage interface {
	SortedSetRangeByScore(ctx context.Context, key, min, max string, offset, count int64) ([]string, error)
	Get(ctx context.Context, key string) (string, error)
	Remove(ctx context.Context, key string) error
	SortedSetRemove(ctx context.Context, set string, value interface{}) error
}

type publisher interface {
	Publish(ctx context.Context, queueName string, value interface{}) error
}

type RedisPoller struct {
	storage   storage
	publisher publisher
}

func NewRedisPoller(storage storage, publisher publisher) *RedisPoller {
	return &RedisPoller{storage: storage, publisher: publisher}
}

func (rp *RedisPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rp.processReadyTasks(ctx)
		}
	}
}
func (rp *RedisPoller) processReadyTasks(ctx context.Context) {
	now := time.Now().UnixMilli()

	taskIDs, err := rp.storage.SortedSetRangeByScore(
		ctx, "delayed:notifications", "-inf", strconv.FormatInt(now, 10), 0, 50)
	if err != nil {
	}

	for _, id := range taskIDs {
		payload, err := rp.storage.Get(ctx, "task:"+id)
		if err != nil {
			continue
		}

		// if err := rp.rabbitPublisher.Publish(ctx, "notifications", payload); err != nil {
		// 	continue
		// }
		if err := rp.publisher.Publish(ctx, "notifications", payload); err != nil {
			continue
		}

		// 
		if err := rp.storage.SortedSetRemove(ctx, "delayed:notifications", id); err != nil {

		}
		if err := rp.storage.Remove(ctx, "task:"+id); err != nil {

		}
	}
}
