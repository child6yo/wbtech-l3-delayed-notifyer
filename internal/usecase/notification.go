package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	"github.com/google/uuid"
)

type storage interface {
	Add(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
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
// Вощврашает айди запланнированного уведомления.
func (nc *NotificationCreator) ScheduleNotification(ctx context.Context, notification models.DelayedNotification) (string, error) {
	uid := uuid.NewString()
	notification.ID = uid

	payload, err := json.Marshal(notification)
	if err != nil {
		return "", err
	}

	err = nc.storage.Add(ctx, "notification:"+notification.ID, payload)
	if err != nil {
		return "", err
	}

	err = nc.storage.Add(ctx, "notification.status:"+notification.ID, string(models.StatusScheduled))
	if err != nil {
		return "", err
	}

	sendAtTimestamp := time.Now().Add(notification.Delay).UnixMilli()
	err = nc.storage.SortedSetAdd(
		ctx, nc.delayedSetName, notification.ID, float64(sendAtTimestamp))
	if err != nil {
		// на данный момент обеспечивает атомарность операции
		// т.е. удаляет payload по ключу в случае ошибки при добавлении в sorted set
		_ = nc.storage.Remove(ctx, "notification:"+notification.ID)
		_ = nc.storage.Remove(ctx, "notification.status:"+notification.ID)
		return "", err
	}

	return uid, nil
}

// GetNotification возвращает уведомление по его айди.
func (nc *NotificationCreator) GetNotification(ctx context.Context, uid string) (models.NotificationStatus, error) {
	notification, err := nc.storage.Get(ctx, "notification.status:"+uid)
	if err != nil {
		return "", err
	}

	return models.NotificationStatus(notification), err
}
