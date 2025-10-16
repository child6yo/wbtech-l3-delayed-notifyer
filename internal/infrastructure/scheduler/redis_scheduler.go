package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	r "github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/redis"
)

type RedisScheduler struct {
	redis       *redis.Client
	delayedList string
}

func NewRedisScheduler(client *redis.Client) *RedisScheduler {
	return &RedisScheduler{redis: client}
}

func (rs *RedisScheduler) ScheduleNotification(ctx context.Context, notification models.DelayedNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = rs.redis.Set(ctx, "task:"+notification.ID, payload)
	if err != nil {
		return err
	}

	sendAtTimestamp := time.Now().Add(notification.Delay).Unix()
	_, err = rs.redis.ZAdd(ctx, rs.delayedList, &r.Z{
		Score:  float64(sendAtTimestamp),
		Member: notification.ID,
	}).Result()

	return err
}
