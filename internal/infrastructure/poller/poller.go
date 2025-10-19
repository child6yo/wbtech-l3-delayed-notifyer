package poller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/logger"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
)

type storage interface {
	SortedSetRangeByScore(ctx context.Context, key, min, max string, offset, count int64) ([]string, error)
	Add(ctx context.Context, key string, value interface{}) error
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
	logger         logger.Logger
}

// NewRedisPoller создает новый RedisPoller.
func NewRedisPoller(storage storage, publisher publisher, delayedSetName string, logger logger.Logger) *RedisPoller {
	return &RedisPoller{
		storage: storage, publisher: publisher, delayedSetName: delayedSetName, logger: logger}
}

// Run запускает поллер. Поллер запускает функцию-воркер с частотой тикера.
func (rp *RedisPoller) Run(ctx context.Context, ticker *time.Ticker) {
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// с каждым тиком запускаем воркер
			// следующий тик будет доступен только если функция завершена (очевидно)
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
		rp.logger.Error(err)
	}

	for _, id := range notificationIDs {
		rp.handleNotification(ctx, id)
	}
}

// обработка уведомления.
func (rp *RedisPoller) handleNotification(ctx context.Context, notificationID string) {
	payload, err := rp.storage.Get(ctx, "notification:"+notificationID)
	if err != nil {
		return
	}

	if err := rp.publisher.Publish(payload); err != nil {
		_ = rp.storage.Add(ctx, "notification.status:"+notificationID, string(models.StatusFailed))
		rp.logger.WithFields("notificationID", notificationID).Error(fmt.Errorf("publishing: %v", err))
		return
	}

	if err := rp.storage.SortedSetRemove(ctx, rp.delayedSetName, notificationID); err != nil {
		rp.logger.WithFields("notificationID", notificationID).Error(err)
	}

	if err := rp.storage.Remove(ctx, "notification:"+notificationID); err != nil {
		rp.logger.WithFields("notificationID", notificationID).Error(err)
	}

	if err := rp.storage.Add(ctx, "notification.status:"+notificationID, string(models.StatusSending)); err != nil {
		rp.logger.WithFields("notificationID", notificationID).Error(err)
	}
}
