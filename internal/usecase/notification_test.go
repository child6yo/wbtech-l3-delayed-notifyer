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

func TestNotificationCreator_ScheduleNotification(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_usecase.NewMockstorage(ctrl)
	creator := NewNotificationCreator(mockStorage, "delayed_notifications")

	notification := models.DelayedNotification{
		Notification: "test message",
		Delay:        10 * time.Second,
		Channels: models.Channels{
			EmailChannel: models.EmailChannel{Email: "user@example.com"},
		},
	}

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil).Times(2)
		mockStorage.EXPECT().SortedSetAdd(gomock.Any(), "delayed_notifications", gomock.Any(), gomock.Any()).Return(nil)

		id, err := creator.ScheduleNotification(context.Background(), notification)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("storage_add_fails_on_payload", func(t *testing.T) {
		mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(errors.New("storage error"))

		_, err := creator.ScheduleNotification(context.Background(), notification)
		assert.Error(t, err)
	})

	t.Run("storage_add_fails_on_status", func(t *testing.T) {
		mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil)
		mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(errors.New("storage error"))

		_, err := creator.ScheduleNotification(context.Background(), notification)
		assert.Error(t, err)
	})

	t.Run("sorted_set_add_fails", func(t *testing.T) {
		mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil).Times(2)
		mockStorage.EXPECT().SortedSetAdd(gomock.Any(), "delayed_notifications", gomock.Any(), gomock.Any()).Return(errors.New("zadd error"))
		mockStorage.EXPECT().Remove(gomock.Any(), gomock.Not(gomock.Nil())).Return(nil).Times(2)

		_, err := creator.ScheduleNotification(context.Background(), notification)
		assert.Error(t, err)
	})
}

func TestNotificationCreator_GetNotificationStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_usecase.NewMockstorage(ctrl)
	creator := NewNotificationCreator(mockStorage, "delayed_notifications")

	t.Run("success", func(t *testing.T) {
		expectedStatus := string(models.StatusScheduled)
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return(expectedStatus, nil)

		status, err := creator.GetNotificationStatus(context.Background(), "test-id")
		require.NoError(t, err)
		assert.Equal(t, models.StatusScheduled, status)
	})

	t.Run("storage_error", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return("", errors.New("storage error"))

		_, err := creator.GetNotificationStatus(context.Background(), "test-id")
		assert.Error(t, err)
	})
}

func TestNotificationCreator_RemoveNotification(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_usecase.NewMockstorage(ctrl)
	creator := NewNotificationCreator(mockStorage, "delayed_notifications")

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return(string(models.StatusScheduled), nil)
		mockStorage.EXPECT().Remove(gomock.Any(), "notification:test-id").Return(nil)
		mockStorage.EXPECT().Remove(gomock.Any(), "notification.status:test-id").Return(nil)

		err := creator.RemoveNotification(context.Background(), "test-id")
		assert.NoError(t, err)
	})

	t.Run("already_sent", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return(string(models.StatusSent), nil)

		err := creator.RemoveNotification(context.Background(), "test-id")
		assert.EqualError(t, err, "notification test-id already sent")
	})

	t.Run("get_status_fails", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return("", errors.New("storage error"))

		err := creator.RemoveNotification(context.Background(), "test-id")
		assert.Error(t, err)
	})

	t.Run("remove_payload_fails", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return(string(models.StatusScheduled), nil)
		mockStorage.EXPECT().Remove(gomock.Any(), "notification:test-id").Return(errors.New("remove error"))

		err := creator.RemoveNotification(context.Background(), "test-id")
		assert.Error(t, err)
	})

	t.Run("remove_status_fails", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "notification.status:test-id").Return(string(models.StatusScheduled), nil)
		mockStorage.EXPECT().Remove(gomock.Any(), "notification:test-id").Return(nil)
		mockStorage.EXPECT().Remove(gomock.Any(), "notification.status:test-id").Return(errors.New("remove error"))

		err := creator.RemoveNotification(context.Background(), "test-id")
		assert.Error(t, err)
	})
}

func TestNotificationCreator_ConcurrentSchedule(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_usecase.NewMockstorage(ctrl)
	creator := NewNotificationCreator(mockStorage, "delayed_notifications")

	notification := models.DelayedNotification{
		Notification: "concurrent test",
		Delay:        5 * time.Second,
		Channels: models.Channels{
			TelegramChannel: models.TelegramChannel{ChatID: "123456"},
		},
	}

	const goroutines = 10
	errCh := make(chan error, goroutines)

	mockStorage.EXPECT().Add(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil).Times(2 * goroutines)
	mockStorage.EXPECT().SortedSetAdd(gomock.Any(), "delayed_notifications", gomock.Any(), gomock.Any()).Return(nil).Times(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := creator.ScheduleNotification(context.Background(), notification)
			errCh <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		err := <-errCh
		assert.NoError(t, err)
	}
}
