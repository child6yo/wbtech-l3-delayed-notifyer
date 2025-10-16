package poller

import "github.com/wb-go/wbf/redis"

type RedisPoller struct {
	redis redis.Client
}

func NewRedisPoller(client redis.Client) *RedisPoller {
	return &RedisPoller{redis: client}
}

