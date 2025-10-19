package messaging

import (
	"errors"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
)

// RabbitMQBroker определяет структуру соединения с RabbitMQ.
type RabbitMQBroker struct {
	url   string // аддресс
	queue string // очередь

	ch *amqp091.Channel

	publisher *rabbitmq.Publisher
	consumer  *rabbitmq.Consumer
}

// NewRabbitMQBroker создает новый RabbitMQBroker.
func NewRabbitMQBroker(url, queue string) *RabbitMQBroker {
	return &RabbitMQBroker{url: url, queue: queue}
}

// ConnectWithRetry n-е кол-во раз пытается подключиться к RabbitMQ.
func (b *RabbitMQBroker) ConnectWithRetry(retries int, pause time.Duration) error {
	conn, err := rabbitmq.Connect(b.url, retries, pause)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	b.ch = ch

	qm := rabbitmq.NewQueueManager(ch)
	_, err = qm.DeclareQueue(b.queue)
	if err != nil {
		return err
	}

	b.publisher = rabbitmq.NewPublisher(ch, "")
	b.consumer = rabbitmq.NewConsumer(ch, rabbitmq.NewConsumerConfig(b.queue))

	return nil
}

// Publish публикует значение в очередь.
func (b *RabbitMQBroker) Publish(value string) error {
	if b.publisher == nil {
		return errors.New("not connected: publisher is nil")
	}
	return b.publisher.Publish([]byte(value), b.queue, "json")
}

// Consume запускает консьюмер, прокидывающий сообщения из очереди в msgChan.
// Ack происходит автоматически внутри consumer.
func (b *RabbitMQBroker) Consume(msgChan chan []byte) error {
	if b.consumer == nil {
		return errors.New("not connected: consumer is nil")
	}

	return b.consumer.Consume(msgChan)
}

// Close закрывает соединение amqp канала.
func (b *RabbitMQBroker) Close() error {
	return b.ch.Close()
}
