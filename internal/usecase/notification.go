package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
)

type storage interface {
	Add(ctx context.Context, key string, value interface{}) error
	Remove(ctx context.Context, key string) error
	SortedSetAdd(ctx context.Context, set string, value interface{}, score float64) error
}

// NotificationCreator отвечает за логику создания новых уведомлений в отложенной очереди.
type NotificationCreator struct {
	storage        storage // место хранения отложенной очереди.
	delayedSetName string  // название очереди
}

// NewNotificationCreator создает новый NotificationCreator.
func NewNotificationCreator(storage storage, delayedSetName string) *NotificationCreator {
	return &NotificationCreator{storage: storage, delayedSetName: delayedSetName}
}

// ScheduleNotification кладет новое уведомление в отложенную очередь.
func (nc *NotificationCreator) ScheduleNotification(ctx context.Context, notification models.DelayedNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	key := "notification:" + notification.ID

	err = nc.storage.Add(ctx, key, payload)
	if err != nil {
		return err
	}

	sendAtTimestamp := time.Now().Add(notification.Delay).UnixMilli()
	err = nc.storage.SortedSetAdd(
		ctx, nc.delayedSetName, notification.ID, float64(sendAtTimestamp))
	if err != nil {
		// на данный момент обеспечивает атомарность операции
		// т.е. удаляет payload по ключу в случае ошибки при добавлении в sorted set
		_ = nc.storage.Remove(ctx, key)
		return err
	}

	return nil
}
