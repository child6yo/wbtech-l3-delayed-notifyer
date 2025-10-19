package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/controller/consumer"
	httpctrl "github.com/child6yo/wbtech-l3-delayed-notifyer/internal/controller/http"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/messaging"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/poller"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/repository"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/sender"
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/usecase"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	createNotificationRoute    = "/notify"
	getNotificationStatusRoute = "/notify/:id"
	deleteNotificationRoute    = "/notify/:id"
)

type appConfig struct {
	address string

	redisAddr             string
	redisPassword         string
	redisDB               int
	redisDelayedQueueName string

	rabbitMQAddr  string
	rabbitMQQueue string

	pollerTick int

	emailFrom string
	emailHost string
	emailPort string

	tgBotToken string
}

func initConfig(configFilePath, envFilePath, envPrefix string) (*appConfig, error) {
	appConfig := &appConfig{}

	cfg := config.New()

	err := cfg.Load(configFilePath, envFilePath, envPrefix)
	if err != nil {
		return appConfig, fmt.Errorf("failed to load config: %w", err)
	}

	appConfig.address = cfg.GetString("app_address")

	appConfig.redisAddr = cfg.GetString("redis_address")
	appConfig.redisPassword = cfg.GetString("redis_password")
	appConfig.redisDB = cfg.GetInt("redis_db")
	appConfig.redisDelayedQueueName = cfg.GetString("redis_delayed_queue")

	appConfig.rabbitMQAddr = cfg.GetString("rabbitmq_address")
	appConfig.rabbitMQQueue = cfg.GetString("rabbitmq_queue")

	appConfig.pollerTick = cfg.GetInt("poller_tick_milliseconds")

	appConfig.emailFrom = cfg.GetString("smtp_from")
	appConfig.emailHost = cfg.GetString("smtp_host")
	appConfig.emailPort = cfg.GetString("smtp_port")

	appConfig.tgBotToken = cfg.GetString("TG_BOT_TOKEN")

	return appConfig, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	zlog.InitConsole()
	lgr := zlog.Logger

	cfg, err := initConfig("config/config.yml", ".env", "")
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}

	rds := repository.NewRedis(cfg.redisAddr, cfg.redisPassword, cfg.redisDB)
	pbl := messaging.NewRabbitMQBroker(cfg.rabbitMQAddr, cfg.rabbitMQQueue)
	if err := pbl.ConnectWithRetry(3, 1*time.Second); err != nil {
		lgr.Fatal().Err(err).Send()
	}

	msgChan := make(chan []byte, 10)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer close(msgChan)
		defer wg.Done()
		if err := pbl.Consume(msgChan); err != nil {
			lgr.Err(err).Send()
		}
	}()

	pl := poller.NewRedisPoller(rds, pbl, cfg.redisDelayedQueueName, poller.NewLoggerAdapter(lgr))
	wg.Add(1)
	go func() {
		defer wg.Done()
		pl.Run(ctx, time.NewTicker(time.Duration(cfg.pollerTick)*time.Millisecond))
	}()

	emailSender := sender.NewEmailSender(cfg.emailFrom, cfg.emailHost, cfg.emailPort)

	tgSender, err := sender.NewTelegramSender(cfg.tgBotToken)
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		tgSender.Start(ctx)
	}()

	ns := usecase.NewNotificationSender(emailSender, tgSender, rds)
	cnsHandler := consumer.NewNotificationConsumer(msgChan, consumer.NewLoggerAdapter(lgr), ns)
	wg.Add(1)
	go func() {
		defer wg.Done()
		cnsHandler.Consume(ctx)
	}()

	nuc := usecase.NewNotificationCreator(rds, cfg.redisDelayedQueueName)
	nc := httpctrl.NewNotificationsController(nuc)
	mdlw := httpctrl.NewMiddleware(httpctrl.NewLoggerAdapter(lgr))

	srv := ginext.New("")
	srv.Use(ginext.Logger(), ginext.Recovery(), mdlw.ErrHandlingMiddleware())
	srv.POST(createNotificationRoute, nc.CreateNotification)
	srv.GET(getNotificationStatusRoute, nc.GetNotificationStatus)
	srv.DELETE(deleteNotificationRoute, nc.DeleteNotification)

	httpServer := &http.Server{
		Addr:    cfg.address,
		Handler: srv,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lgr.Err(err).Send()
		}
	}()

	<-ctx.Done()
	lgr.Info().Msg("shutting down gracefully...")

	if err := pbl.Close(); err != nil {
		lgr.Err(err).Send()
	}

	if err := httpServer.Shutdown(context.Background()); err != nil {
		lgr.Err(err).Send()
	}

	if err := tgSender.Stop(context.Background()); err != nil {
		lgr.Err(err).Send()
	}

	wg.Wait()

	lgr.Info().Msg("app exited")
}
