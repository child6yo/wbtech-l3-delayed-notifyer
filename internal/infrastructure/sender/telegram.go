package sender

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Telegram определяет отправщик сообщений через телеграм-канал.
type Telegram struct {
	Bot *bot.Bot
}

// NewTelegram создает новый Telegram.
func NewTelegram(token string) (*Telegram, error) {
	b, err := bot.New(token)

	return &Telegram{Bot: b}, err
}

// Start запускает работу телеграм отправщика.
func (t *Telegram) Start(ctx context.Context) {
	t.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text: fmt.Sprintf(`Ваш чат айди - %d. 
			Используйте его, если хотите получать уведомления через телеграм-канал.`, chatID),
		})
		if err != nil {
			return
		}
	})

	t.Bot.Start(ctx)
}

// Send отправляет сообщение на указанный chatID.
func (t *Telegram) Send(chatID string, data string) error {
	_, err := t.Bot.SendMessage(context.Background(), &bot.SendMessageParams{ChatID: chatID, Text: data})
	return err
}

// Stop делает попытку закрыть соединение.
func (t *Telegram) Stop(ctx context.Context) error {
	_, err := t.Bot.Close(ctx)
	return err
}
