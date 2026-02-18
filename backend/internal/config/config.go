package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Server
	Port        string `env:"PORT,default=8080"`
	Env         string `env:"ENV,default=development"`
	FrontendURL string `env:"FRONTEND_URL"`
	// Database
	DB DBConfig
	// JWT
	JWTSecret string `env:"JWT_SECRET,required"`

	// RabbitMQ
	RabbitMQ RabbitMQConfig

	// Redis
	Redis RedisConfig

	// Bank URLs
	BankURLs BankURLsConfig

	// Logging
	LogLevel  string `env:"LOG_LEVEL,default=info"`
	LogFormat string `env:"LOG_FORMAT,default=json"`

	// CORS
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS"`

	// Application
	MaxRetries     int           `env:"MAX_RETRIES,default=3"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT,default=30s"`
	IdempotencyTTL time.Duration `env:"IDEMPOTENCY_TTL,default=24h"`
}

type DBConfig struct {
	Server          string `env:"POSTGRES_SERVER,required"`
	Database        string `env:"POSTGRES_DATABASE,required"`
	Port            int    `env:"POSTGRES_PORT,default=5432"`
	User            string `env:"POSTGRES_USER,required"`
	Password        string `env:"POSTGRES_PASSWORD,required"`
	ConnectTimeOut  int    `env:"POSTGRES_CONNECT_TIMEOUT,default=15"`
	MaxOpenConns    int    `env:"POSTGRES_MAX_OPEN_CONNS,default=30"`
	MaxIdleConns    int    `env:"POSTGRES_MAX_IDLE_CONNS,default=25"`
	ConnMaxLifetime int    `env:"POSTGRES_CONN_MAX_LIFETIME,default=30"`
	QueryTimeout    int    `env:"POSTGRES_QUERY_TIMEOUT,default=60"`
}

type RabbitMQConfig struct {
	Host     string `env:"RABBITMQ_HOST,required"`
	Port     int    `env:"RABBITMQ_PORT,default=5672"`
	User     string `env:"RABBITMQ_USER,required"`
	Password string `env:"RABBITMQ_PASSWORD,required"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	Port     int    `env:"REDIS_PORT,default=6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB,default=0"`
}

type BankURLsConfig struct {
	ESP string `env:"ESP_BANK_URL,default=http://localhost:8080/esp"`
	PT  string `env:"PT_BANK_URL,default=http://localhost:8080/pt"`
	IT  string `env:"IT_BANK_URL,default=http://localhost:8080/it"`
	MX  string `env:"MX_BANK_URL,default=http://localhost:8080/mx"`
	CO  string `env:"CO_BANK_URL,default=http://localhost:8080/co"`
	BR  string `env:"BR_BANK_URL,default=http://localhost:8080/br"`
}

func Load() *Config {
	ctx := context.Background()
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		panic(fmt.Sprintf("Error loading config: %v", err))
	}
	return &cfg
}