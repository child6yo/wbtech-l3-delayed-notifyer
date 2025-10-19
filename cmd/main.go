package main

import (
	"context"
	"log"
	"time"

	httpctrl "github.com/child6yo/wbtech-l3-delayed-notifyer/internal/controller/http"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/poller"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/publisher"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/repository"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/usecase"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	createNotificationRoute    = "/notify"
	getNotificationStatusRoute = "/notify/:id"
	deleteNotificationRoute    = "/notify/:id"
)

func main() {
	zlog.InitConsole()
	lgr := zlog.Logger

	rds := repository.NewRedis("localhost:6380", "", 0)
	pbl := publisher.NewRabbitMQPublisher("amqp://localhost:5672", "notification.created")
	err := pbl.ConnectWithRetry(3, 1*time.Second)
	if err != nil {
		log.Println(err)
	}

	pl := poller.NewRedisPoller(rds, pbl, "delayed_queue", poller.NewLoggerAdapter(lgr))
	go func() {
		pl.Run(context.Background(), time.NewTicker(100*time.Millisecond))
	}()

	nuc := usecase.NewNotificationCreator(rds, "delayed_queue")
	nc := httpctrl.NewNotificationsController(nuc)
	mdlw := httpctrl.NewMiddleware(httpctrl.NewLoggerAdapter(lgr))

	srv := ginext.New("")
	srv.Use(ginext.Logger(), ginext.Recovery(), mdlw.ErrHandlingMiddleware())
	srv.POST(createNotificationRoute, nc.CreateNotification)
	srv.GET(getNotificationStatusRoute, nc.GetNotificationStatus)
	srv.DELETE(deleteNotificationRoute, nc.DeleteNotification)
	srv.Run("localhost:8080")
}
