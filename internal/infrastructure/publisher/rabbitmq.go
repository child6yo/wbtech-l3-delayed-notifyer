package publisher

import (
	"time"

	"github.com/wb-go/wbf/rabbitmq"
)

type RabbitMQPublisher struct {
	url   string
	queue string

	publisher *rabbitmq.Publisher
}

func NewRabbitMQPublisher(url, queueName string) *RabbitMQPublisher {
	return &RabbitMQPublisher{url: url, queue: queueName}
}

func (p *RabbitMQPublisher) ConnectWithRetry(retries int, pause time.Duration) error {
	conn, err := rabbitmq.Connect(p.url, retries, pause)
	if err != nil {
		return err
	}

	c, err := conn.Channel()
	if err != nil {
		return err
	}

	q := rabbitmq.NewQueueManager(c)
	_, err = q.DeclareQueue(p.queue)
	if err != nil {
		return err
	}

	publisher := rabbitmq.NewPublisher(c, "")
	p.publisher = publisher

	return nil
}

func (p *RabbitMQPublisher) Publish(value string) error {
	return p.publisher.Publish([]byte(value), p.queue, "json")
}
