package consumer

import (
	"context"
	"encoding/json"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
)

type notificationSender interface {
	Send(notification models.DelayedNotification) error
}

type logger interface {
	WithFields(keyValues ...interface{}) logger
	Error(err error)
	Debug(msg string)
}

// NotificationConsumer принимает уведомления через канал и рассылает их.
type NotificationConsumer struct {
	usecase notificationSender

	msgChan chan []byte
	logger  logger
}

// NewNotificationConsumer создает новый NotificationConsumer.
func NewNotificationConsumer(msgChan chan []byte, logger logger, uc notificationSender) *NotificationConsumer {
	return &NotificationConsumer{usecase: uc, msgChan: msgChan, logger: logger}
}

// Consume обрабатывает канал сообщений.
func (c *NotificationConsumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.msgChan:
			if len(msg) == 0 || !ok {
				continue
			}

			c.logger.WithFields("msg", string(msg)).Debug("new msg consumed")

			var notification models.DelayedNotification
			err := json.Unmarshal(msg, &notification)
			if err != nil {
				c.logger.WithFields("data", msg).Error(err)
				continue
			}

			if err := c.usecase.Send(notification); err != nil {
				c.logger.WithFields("notification", notification).Error(err)
			}
		}
	}
}
