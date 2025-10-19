package usecase

import (
	"context"
	"errors"
	"sync"
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

// Send отправляет уведомления по указанным каналам и сохраняет статус.
func (ns *NotificationSender) Send(ctx context.Context, notification models.DelayedNotification) error {
	errs := ns.sendNotifications(ctx, notification)

	status := ns.determineStatus(errs)
	if err := ns.saveStatus(ctx, notification.ID, status); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// sendNotifications отправляет уведомления по email и Telegram (если указаны) с retry-логикой.
func (ns *NotificationSender) sendNotifications(ctx context.Context, notification models.DelayedNotification) []error {
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	if email := notification.Channels.EmailChannel.Email; email != "" {
		wg.Add(1)
		go ns.sendWithRetry(ctx, &wg, errCh, func() error {
			return ns.emailSender.Send(email, string(notification.Notification))
		})
	}

	if tg := notification.Channels.TelegramChannel.ChatID; tg != "" {
		wg.Add(1)
		go ns.sendWithRetry(ctx, &wg, errCh, func() error {
			return ns.tgSender.Send(tg, string(notification.Notification))
		})
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errs
}

// sendWithRetry оборачивает отправку с механизмом повторных попыток.
func (ns *NotificationSender) sendWithRetry(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, sendFunc func() error) {
	defer wg.Done()
	err := retry.Do(sendFunc,
		retry.Strategy{
			Attempts: 10,
			Delay:    2 * time.Second,
			Backoff:  2,
		},
	)
	if err != nil {
		select {
		case errCh <- err:
		case <-ctx.Done():
		}
	}
}

// determineStatus определяет статус уведомления на основе наличия ошибок.
func (ns *NotificationSender) determineStatus(errs []error) models.NotificationStatus {
	if len(errs) > 0 {
		return models.StatusFailed
	}
	return models.StatusSent
}

// saveStatus сохраняет статус уведомления в хранилище.
func (ns *NotificationSender) saveStatus(ctx context.Context, notificationID string, status models.NotificationStatus) error {
	return ns.storageAdder.Add(ctx, "notification.status:"+notificationID, string(status))
}
