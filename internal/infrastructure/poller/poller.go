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
	Publish(value string) error
}

// RedisPoller мониторит хранилище уведомлений в поисках тех, которые пора отправить.
// Отправляет необходимые уведомления в паблишер. Пишет ошибки в отдельный канал.
type RedisPoller struct {
	storage        storage
	publisher      publisher
	delayedSetName string
	errCh          chan<- error
}

// NewRedisPoller создает новый RedisPoller.
func NewRedisPoller(storage storage, publisher publisher, delayedSetName string, errCh chan<- error) *RedisPoller {
	return &RedisPoller{
		storage: storage, publisher: publisher, delayedSetName: delayedSetName, errCh: errCh}
}

// Run запускает поллер. Поллер запускает функцию-воркер с частотой тикера.
func (rp *RedisPoller) Run(ctx context.Context, ticker *time.Ticker) {
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

// мониторинговая функция-воркер
func (rp *RedisPoller) processReadyTasks(ctx context.Context) {
	now := time.Now().UnixMilli()

	notificationIDs, err := rp.storage.SortedSetRangeByScore(
		ctx, rp.delayedSetName, "-inf", strconv.FormatInt(now, 10), 0, 10)
	if err != nil {
		rp.errCh <- err
	}

	for _, id := range notificationIDs {
		rp.handleNotification(ctx, id)
	}
}

// обработка уведомления.
func (rp *RedisPoller) handleNotification(ctx context.Context, notificationID string) {
	payload, _ := rp.storage.Get(ctx, notificationID)

	if err := rp.publisher.Publish(payload); err != nil {
		rp.errCh <- err
	}

	if err := rp.storage.SortedSetRemove(ctx, rp.delayedSetName, notificationID); err != nil {
		rp.errCh <- err
	}

	if err := rp.storage.Remove(ctx, notificationID); err != nil {
		rp.errCh <- err
	}
}
