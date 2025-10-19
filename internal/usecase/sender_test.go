package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	mock_usecase "github.com/child6yo/wbtech-l3-delayed-notifyer/internal/usecase/mock"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/pkg/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationSender_Send(t *testing.T) {
	t.Parallel()

	const (
		testID      = "notif-123"
		testMessage = "Hello, world!"
	)

	tests := []struct {
		name           string
		emailAddr      string
		chatID         string
		emailSendErr   error
		tgSendErr      error
		statusSaveErr  error
		expectStatus   models.NotificationStatus
		expectErr      bool
		emailCallCount int
		tgCallCount    int
	}{
		{
			name:           "both channels succeed",
			emailAddr:      "user@example.com",
			chatID:         "123456",
			expectStatus:   models.StatusSent,
			emailCallCount: 1,
			tgCallCount:    1,
		},
		{
			name:           "email fails, telegram succeeds",
			emailAddr:      "user@example.com",
			chatID:         "123456",
			emailSendErr:   errors.New("smtp timeout"),
			expectStatus:   models.StatusFailed,
			expectErr:      true,
			emailCallCount: 3, // retry 3 times
			tgCallCount:    1,
		},
		{
			name:           "telegram fails, email succeeds",
			emailAddr:      "user@example.com",
			chatID:         "123456",
			tgSendErr:      errors.New("tg api down"),
			expectStatus:   models.StatusFailed,
			expectErr:      true,
			emailCallCount: 1,
			tgCallCount:    3,
		},
		{
			name:           "both fail",
			emailAddr:      "user@example.com",
			chatID:         "123456",
			emailSendErr:   errors.New("email error"),
			tgSendErr:      errors.New("tg error"),
			expectStatus:   models.StatusFailed,
			expectErr:      true,
			emailCallCount: 3,
			tgCallCount:    3,
		},
		{
			name:           "email only, success",
			emailAddr:      "user@example.com",
			expectStatus:   models.StatusSent,
			emailCallCount: 1,
		},
		{
			name:         "telegram only, success",
			chatID:       "123456",
			expectStatus: models.StatusSent,
			tgCallCount:  1,
		},
		{
			name:         "no channels",
			expectStatus: models.StatusSent,
		},
		{
			name:           "status save fails",
			emailAddr:      "user@example.com",
			statusSaveErr:  errors.New("redis down"),
			expectStatus:   models.StatusSent,
			expectErr:      true,
			emailCallCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEmail := mock_usecase.NewMockemailSender(ctrl)
			mockTg := mock_usecase.NewMocktelegramSender(ctrl)
			mockStorage := mock_usecase.NewMockstorageAdder(ctrl)

			if tt.emailAddr != "" {
				mockEmail.EXPECT().
					Send(tt.emailAddr, testMessage).
					Return(tt.emailSendErr).
					Times(tt.emailCallCount)
			}

			if tt.chatID != "" {
				mockTg.EXPECT().
					Send(tt.chatID, testMessage).
					Return(tt.tgSendErr).
					Times(tt.tgCallCount)
			}

			mockStorage.EXPECT().
				Add(gomock.Any(), "notification.status:"+testID, string(tt.expectStatus), 168*time.Hour).
				Return(tt.statusSaveErr)

			sender := NewNotificationSender(
				mockEmail,
				mockTg,
				mockStorage,
				3, // retry attempts
				10*time.Millisecond,
				1.0, // no backoff for test speed
			)

			notification := models.DelayedNotification{
				ID:           testID,
				Notification: models.Notification(testMessage),
				Channels: models.Channels{
					EmailChannel:    models.EmailChannel{Email: tt.emailAddr},
					TelegramChannel: models.TelegramChannel{ChatID: tt.chatID},
				},
			}

			err := sender.Send(context.Background(), notification)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationSender_Send_StatusSavedEvenWithContextCancel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEmail := mock_usecase.NewMockemailSender(ctrl)
	mockTg := mock_usecase.NewMocktelegramSender(ctrl)
	mockStorage := mock_usecase.NewMockstorageAdder(ctrl)

	mockEmail.EXPECT().Send("user@example.com", "msg").Return(nil)
	mockStorage.EXPECT().
		Add(gomock.Any(), "notification.status:test", string(models.StatusSent), 168*time.Hour).
		Return(nil)

	sender := NewNotificationSender(mockEmail, mockTg, mockStorage, 1, 0, 1.0)

	notification := models.DelayedNotification{
		ID:           "test",
		Notification: "msg",
		Channels: models.Channels{
			EmailChannel: models.EmailChannel{Email: "user@example.com"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.Send(ctx, notification)
	require.NoError(t, err)
}

func TestNotificationSender_Send_RetryBehavior(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEmail := mock_usecase.NewMockemailSender(ctrl)
	mockTg := mock_usecase.NewMocktelegramSender(ctrl)
	mockStorage := mock_usecase.NewMockstorageAdder(ctrl)

	// Fail twice, succeed on third attempt
	gomock.InOrder(
		mockEmail.EXPECT().Send("test@example.com", "retry me").Return(errors.New("temp fail")),
		mockEmail.EXPECT().Send("test@example.com", "retry me").Return(errors.New("temp fail")),
		mockEmail.EXPECT().Send("test@example.com", "retry me").Return(nil),
	)

	mockStorage.EXPECT().
		Add(gomock.Any(), "notification.status:test", string(models.StatusSent), 168*time.Hour).
		Return(nil)

	sender := NewNotificationSender(
		mockEmail,
		mockTg,
		mockStorage,
		3,
		1*time.Millisecond,
		1.0,
	)

	notification := models.DelayedNotification{
		ID:           "test",
		Notification: "retry me",
		Channels: models.Channels{
			EmailChannel: models.EmailChannel{Email: "test@example.com"},
		},
	}

	err := sender.Send(context.Background(), notification)
	require.NoError(t, err)
}

func TestNotificationSender_Send_NoChannels(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_usecase.NewMockstorageAdder(ctrl)
	mockStorage.EXPECT().
		Add(gomock.Any(), "notification.status:empty", string(models.StatusSent), 168*time.Hour).
		Return(nil)

	sender := NewNotificationSender(
		mock_usecase.NewMockemailSender(ctrl),
		mock_usecase.NewMocktelegramSender(ctrl),
		mockStorage,
		1, 0, 1.0,
	)

	notification := models.DelayedNotification{
		ID:           "empty",
		Notification: "no channels",
		Channels:     models.Channels{},
	}

	err := sender.Send(context.Background(), notification)
	require.NoError(t, err)
}
