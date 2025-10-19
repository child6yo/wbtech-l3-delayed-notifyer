package consumer

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/logger"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
)

type notificationSender interface {
	Send(ctx context.Context, notification models.DelayedNotification) error
}

// NotificationConsumer принимает уведомления через канал и рассылает их.
type NotificationConsumer struct {
	usecase notificationSender

	msgChan chan []byte
	logger  logger.Logger
}

// NewNotificationConsumer создает новый NotificationConsumer.
func NewNotificationConsumer(msgChan chan []byte, logger logger.Logger, uc notificationSender) *NotificationConsumer {
	return &NotificationConsumer{usecase: uc, msgChan: msgChan, logger: logger}
}

// Consume обрабатывает канал сообщений.
// Работает конкурретно в n-е количество воркеров.
func (c *NotificationConsumer) Consume(ctx context.Context, numWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-c.msgChan:
					if !ok {
						return
					}
					if len(msg) == 0 {
						continue
					}

					c.logger.WithFields("msg", string(msg)).Debug("new msg consumed")

					var notification models.DelayedNotification
					if err := json.Unmarshal(msg, &notification); err != nil {
						c.logger.WithFields("data", string(msg)).Error(err)
						continue
					}

					if err := c.usecase.Send(ctx, notification); err != nil {
						c.logger.WithFields("notification", notification).Error(err)
					}
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
}
