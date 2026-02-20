package rabbitmq

import (
	"fmt"
	"time"

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

	const maxAttempts = 10
	backoff := 2 * time.Second

	var conn *amqp.Connection
	var err error
	for i := range maxAttempts {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		if i < maxAttempts-1 {
			time.Sleep(backoff)
			backoff = min(backoff*2, 30*time.Second)
		}
	}
	return nil, fmt.Errorf("connect to rabbitmq: %w", err)
}
