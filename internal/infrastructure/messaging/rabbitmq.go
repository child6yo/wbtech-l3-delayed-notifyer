package messaging

import (
	"errors"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
)

type RabbitMQBroker struct {
	url   string
	queue string

	publisher *rabbitmq.Publisher
	consumer  *rabbitmq.Consumer
}

func NewRabbitMQBroker(url, queue string) *RabbitMQBroker {
	return &RabbitMQBroker{url: url, queue: queue}
}

func (b *RabbitMQBroker) ConnectWithRetry(retries int, pause time.Duration) error {
	conn, err := rabbitmq.Connect(b.url, retries, pause)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	qm := rabbitmq.NewQueueManager(ch)
	_, err = qm.DeclareQueue(b.queue)
	if err != nil {
		return err
	}

	b.publisher = rabbitmq.NewPublisher(ch, "")
	b.consumer = rabbitmq.NewConsumer(ch, rabbitmq.NewConsumerConfig(b.queue))

	return nil
}

func (b *RabbitMQBroker) Publish(value string) error {
	if b.publisher == nil {
		return errors.New("not connected: publisher is nil")
	}
	return b.publisher.Publish([]byte(value), b.queue, "json")
}

func (b *RabbitMQBroker) Consume(msgChan chan []byte) error {
	if b.consumer == nil {
		return errors.New("not connected: consumer is nil")
	}

	return b.consumer.Consume(msgChan)
}
