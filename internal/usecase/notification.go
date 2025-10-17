package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
)

type storage interface {
	Add(ctx context.Context, key string, value interface{}) error
	SortedSetAdd(ctx context.Context, set string, value interface{}, score float64) error
}

type NotificationCreator struct {
	storage     storage
	delayedList string
}

func NewNotificationCreator(storage storage) *NotificationCreator {
	return &NotificationCreator{storage: storage}
}

func (nc *NotificationCreator) ScheduleNotification(ctx context.Context, notification models.DelayedNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = nc.storage.Add(ctx, "task:"+notification.ID, payload)
	if err != nil {
		return err
	}

	sendAtTimestamp := time.Now().Add(notification.Delay).Unix()
	err = nc.storage.SortedSetAdd(
		ctx, nc.delayedList, notification.ID, float64(sendAtTimestamp))

	return err
}
