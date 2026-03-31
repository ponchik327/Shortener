package config

import (
	"fmt"
	"time"

	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
)

// Config содержит всю конфигурацию приложения, загружаемую из config.yaml.
type Config struct {
	App       AppConfig       `yaml:"app"`
	Server    ServerConfig    `yaml:"server"`
	Logger    LoggerConfig    `yaml:"logger"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	Shortener ShortenerConfig `yaml:"shortener"`
}

// AppConfig содержит общие метаданные приложения.
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME" validate:"required"`
	Env  string `yaml:"env"  env:"APP_ENV"  validate:"required,oneof=local dev prod"`
}

// ServerConfig содержит конфигурацию HTTP-сервера.
type ServerConfig struct {
	Host         string        `yaml:"host"          env:"SERVER_HOST"          validate:"required"`
	Port         int           `yaml:"port"          env:"SERVER_PORT"          validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `yaml:"read_timeout"  env:"SERVER_READ_TIMEOUT"  validate:"required"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" validate:"required"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"  env:"SERVER_IDLE_TIMEOUT"  validate:"required"`
}

// LoggerConfig содержит конфигурацию логгера.
type LoggerConfig struct {
	Engine string `yaml:"engine" env:"LOGGER_ENGINE" validate:"required,oneof=zap slog zerolog logrus"`
	Level  string `yaml:"level"  env:"LOGGER_LEVEL"  validate:"required"`
}

// DatabaseConfig содержит DSN для подключения к PostgreSQL.
type DatabaseConfig struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN" validate:"required"`
}

// RedisConfig содержит параметры подключения к Redis и настройки кэша.
type RedisConfig struct {
	Address  string        `yaml:"address"   env:"REDIS_ADDRESS"   validate:"required"`
	Password string        `yaml:"password"  env:"REDIS_PASSWORD"`
	CacheTTL time.Duration `yaml:"cache_ttl" env:"REDIS_CACHE_TTL" validate:"required"`
}

// ShortenerConfig содержит настройки бизнес-логики сервиса сокращения ссылок.
type ShortenerConfig struct {
	CodeLength int `yaml:"code_length" env:"SHORTENER_CODE_LENGTH" validate:"required,min=4,max=32"`
}

// Load читает и валидирует конфигурацию из указанного YAML-файла.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenvport.LoadPath(path, cfg); err != nil {
		return nil, fmt.Errorf("load config from %s: %w", path, err)
	}

	return cfg, nil
}
