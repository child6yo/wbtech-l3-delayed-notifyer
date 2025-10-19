package httpctrl

import (
	"context"
	"fmt"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	"github.com/wb-go/wbf/ginext"
)

type notificationUsecase interface {
	ScheduleNotification(ctx context.Context, notification models.DelayedNotification) (string, error)
	GetNotificationStatus(ctx context.Context, uid string) (models.NotificationStatus, error)
	RemoveNotification(ctx context.Context, uid string) error
}

// NotificationsController http контроллер сервиса отложенных уведомлений.
type NotificationsController struct {
	usecase notificationUsecase
}

// NewNotificationsController создает новый NotificationsController.
func NewNotificationsController(uc notificationUsecase) *NotificationsController {
	return &NotificationsController{usecase: uc}
}

type createNotificationRequest struct {
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

	c.Set("request", req)

	delayedNotif := models.DelayedNotification{
		Notification: models.Notification(req.Notification),
		Delay:        time.Duration(req.DelaySeconds) * time.Second,
		Channels:     req.Channels,
	}

	uid, err := nc.usecase.ScheduleNotification(c.Request.Context(), delayedNotif)
	if err != nil {
		c.JSON(500, ginext.H{"error": "failed to schedule notification"})
		c.Error(fmt.Errorf("scheduling failed: %w", err))
		return
	}

	c.JSON(201, ginext.H{"uid": uid})
}

// GetNotification обрабатывает GET /notify/{id} — получение статуса уведомления.
func (nc *NotificationsController) GetNotificationStatus(c *ginext.Context) {
	uid := c.Param("id")
	c.Set("request", uid)

	status, err := nc.usecase.GetNotificationStatus(c.Request.Context(), uid)
	if err != nil {
		c.JSON(500, ginext.H{"error": "failed to get notification"})
		c.Error(fmt.Errorf("get notification failed: %w", err))
		return
	}

	c.JSON(200, ginext.H{"status": status})
}

// DeleteNotification обрабатывает DELETE /notify/{id} — отмена запланированного уведомления.
func (nc *NotificationsController) DeleteNotification(c *ginext.Context) {
	uid := c.Param("id")
	c.Set("request", uid)

	err := nc.usecase.RemoveNotification(c.Request.Context(), uid)
	if err != nil {
		c.JSON(500, ginext.H{"error": "failed to delete notification: " + err.Error()})
		c.Error(fmt.Errorf("delete notification failed: %w", err))
		return
	}

	c.JSON(200, ginext.H{"message": "notification deleted"})
}
