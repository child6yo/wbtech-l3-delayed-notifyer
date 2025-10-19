package sender

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TelegramSender определяет отправщик сообщений через телеграм-канал.
type TelegramSender struct {
	Bot *bot.Bot
}

// NewTelegramSender создает новый TelegramSender.
func NewTelegramSender(token string) (*TelegramSender, error) {
	b, err := bot.New(token)

	return &TelegramSender{Bot: b}, err
}

// Start запускает работу телеграм отправщика.
func (tb *TelegramSender) Start(ctx context.Context) {
	tb.Bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text: fmt.Sprintf(`Ваш чат айди - %d. 
			Используйте его, если хотите получать уведомления через телеграм-канал.`, chatID),
		})
	})

	tb.Bot.Start(ctx)
}

// Send отправляет сообщение на указанный chatID.
func (ts *TelegramSender) Send(chatID string, data string) error {
	_, err := ts.Bot.SendMessage(context.Background(), &bot.SendMessageParams{ChatID: chatID, Text: data})
	return err
}
