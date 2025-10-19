package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	"github.com/wb-go/wbf/retry"
)

type storageAdder interface {
	Add(ctx context.Context, key string, value interface{}) error
}

type telegramSender interface {
	Send(chatID string, data string) error
}

type emailSender interface {
	Send(emailAddr string, data string) error
}

// NotificationSender рассылает уведомления по разным каналам их отправщиками.
type NotificationSender struct {
	emailSender  emailSender
	tgSender     telegramSender
	storageAdder storageAdder
}

// NewNotificationSender создает новый NotificationSender.
func NewNotificationSender(
	emailSender emailSender, tgSender telegramSender, storageAdder storageAdder) *NotificationSender {
	return &NotificationSender{
		emailSender:  emailSender,
		tgSender:     tgSender,
		storageAdder: storageAdder,
	}
}

// Send отправляет уведомления по тем каналам, которые в нем указаны.
func (ns *NotificationSender) Send(notification models.DelayedNotification) error {
	var errs []error

	if email := notification.Channels.EmailChannel.Email; email != "" {
		go func() {
			retry.Do(func() error {
				return ns.emailSender.Send(email, string(notification.Notification))
			}, retry.Strategy{Attempts: 10, Delay: 2 * time.Second, Backoff: 2})
		}()
	}

	if tg := notification.Channels.TelegramChannel.ChatID; tg != "" {
		go func() {
			retry.Do(func() error {
				return ns.tgSender.Send(tg, string(notification.Notification))
			}, retry.Strategy{Attempts: 10, Delay: 2 * time.Second, Backoff: 2})
		}()
	}

	ns.storageAdder.Add(
		context.Background(), "notification.status:"+notification.ID, models.StatusSent)

	return errors.Join(errs...)
}
