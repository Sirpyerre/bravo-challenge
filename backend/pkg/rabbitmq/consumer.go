package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) *Consumer {
	return &Consumer{conn: conn}
}

// Consume declara el exchange, crea y vincula la queue, y empieza a consumir.
// handler recibe el body del mensaje; si retorna error hace Nack (sin requeue).
func (c *Consumer) Consume(queue, routingKey string, handler func([]byte) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}

	if err := ch.QueueBind(q.Name, routingKey, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", queue, err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", queue, err)
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}
