package models

import "time"

type Notification string

type TelegramChannel struct {
	ChatID string `json:"chat_id"`
}

type EmailChannel struct {
	Email string `json:"email"`
}

type Channels struct {
	TelegramChannel TelegramChannel `json:"tg_channel,omitempty"`
	EmailChannel    EmailChannel    `json:"email_channel,omitempty"`
}

type DelayedNotification struct {
	ID           string
	Notification Notification  `json:"notification"`
	Delay        time.Duration `json:"delay"`
	Channels     Channels      `json:"channels"`
}
