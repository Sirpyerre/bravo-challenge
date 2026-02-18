package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
}

func New(cfg Config) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	return conn, nil
}
