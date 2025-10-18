package httpctrl

import (
	"context"
	"fmt"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	"github.com/wb-go/wbf/ginext"
)

type notificationUsecase interface {
	ScheduleNotification(ctx context.Context, notification models.DelayedNotification) error
}

type NotificationsController struct {
	usecase notificationUsecase
}

func NewNotificationsController(uc notificationUsecase) *NotificationsController {
	return &NotificationsController{usecase: uc}
}

type createNotificationRequest struct {
	ID           string          `json:"id" binding:"required,alphanum,min=1,max=64"`
	Notification string          `json:"notification" binding:"required,min=1,max=1000"`
	DelaySeconds int64           `json:"delay_seconds" binding:"required,min=1,max=2592000"` // 1 сек – 30 дней
	Channels     models.Channels `json:"channels" binding:"required"`
}

// CreateNotification обрабатывает POST /notify — создание уведомлений с датой и временем отправки.
func (nc *NotificationsController) CreateNotification(c *ginext.Context) {
	var req createNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ginext.H{"error": "invalid request: " + err.Error()})
		c.Error(fmt.Errorf("validation error: %w", err))
		return
	}

	delayedNotif := models.DelayedNotification{
		ID:           req.ID,
		Notification: models.Notification(req.Notification),
		Delay:        time.Duration(req.DelaySeconds) * time.Second,
		Channels:     req.Channels,
	}

	if err := nc.usecase.ScheduleNotification(c.Request.Context(), delayedNotif); err != nil {
		c.JSON(500, ginext.H{"error": "failed to schedule notification"})
		c.Error(fmt.Errorf("scheduling failed: %w", err))
		return
	}

	c.JSON(201, ginext.H{"message": "notification scheduled"})
}
