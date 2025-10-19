package models

import "time"

// Notification описание уведомления.
type Notification string

// NotificationStatus статус уведомления.
// Использовать ТОЛЬКО с описанными статусами.
type NotificationStatus string

const (
	// StatusScheduled - уведомление запланированно.
	StatusScheduled NotificationStatus = "scheduled"

	// StatusSending - уведомление отправляется.
	StatusSending NotificationStatus = "sending"

	// StatusSent - уведомление отправлено.
	StatusSent NotificationStatus = "sent"

	// StatusFailed - ошибка отправки уведомления.
	StatusFailed NotificationStatus = "failed"
)

// TelegramChannel канал отправки через телеграм.
type TelegramChannel struct {
	ChatID string `json:"chat_id"`
}

// EmailChannel канал отправки через email.
type EmailChannel struct {
	Email string `json:"email"`
}

// Channels определяет возможные каналы отправки уведомления.
type Channels struct {
	TelegramChannel TelegramChannel `json:"tg_channel,omitempty"`
	EmailChannel    EmailChannel    `json:"email_channel,omitempty"`
}

// DelayedNotification определяет модель отложенного уведомления.
type DelayedNotification struct {
	ID           string        `json:"id"`
	Notification Notification  `json:"notification"`
	Delay        time.Duration `json:"delay"`
	Channels     Channels      `json:"channels"`
}
