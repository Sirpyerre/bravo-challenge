package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ExchangeName = "bravo.events"

type Publisher struct {
	conn *amqp.Connection
}

func NewPublisher(conn *amqp.Connection) *Publisher {
	return &Publisher{conn: conn}
}

// Publish envía un mensaje al exchange con el routing key dado.
// Declara el exchange si no existe.
func (p *Publisher) Publish(routingKey string, body []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	return ch.Publish(ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}