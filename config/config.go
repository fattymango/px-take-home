package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

type Logger struct {
	LogFile string `envconfig:"LOG_FILE" validate:"required"`
}

type Server struct {
	Port string `envconfig:"SERVER_PORT" validate:"required,numeric"`
}

type Debug struct {
	Debug bool `envconfig:"DEBUG"`
}

type DB struct {
	Host            string `envconfig:"DB_HOST" validate:"required"`
	Port            string `envconfig:"DB_PORT" validate:"required"`
	User            string `envconfig:"DB_USER" validate:"required"`
	Password        string `envconfig:"DB_PASSWORD" validate:"required"`
	Name            string `envconfig:"DB_NAME" validate:"required"`
	File            string `envconfig:"DB_FILE" validate:"required"` // SQLite file path
	SSLMode         string `envconfig:"DB_SSL_MODE" validate:"required"`
	MaxIdleConns    int    `envconfig:"DB_MAX_IDLE_CONNS" default:"2"`
	MaxOpenConns    int    `envconfig:"DB_MAX_OPEN_CONNS" default:"5"`
	MaxConnLifetime int    `envconfig:"DB_MAX_CONN_LIFETIME" default:"10"`
}

type Redis struct {
	Host string `envconfig:"REDIS_HOST" validate:"required"`
	Port string `envconfig:"REDIS_PORT" validate:"required"`
}

type Swagger struct {
	FilePath string `envconfig:"SWAGGER_FILE_PATH" validate:"required"`
}

type CMD struct {
	Validate bool `envconfig:"CMD_VALIDATE" default:"true"`
}

type TaskLogger struct {
	DirPath string `envconfig:"TASK_LOGGER_DIR_PATH" default:"./task_logs"`
}

type Config struct {
	DB         DB
	Logger     Logger
	Server     Server
	Debug      Debug
	Redis      Redis
	Swagger    Swagger
	CMD        CMD
	TaskLogger TaskLogger
}

func NewConfig() (*Config, error) {
	fmt.Println("Loading configuration...")

	debug := os.Getenv("DEBUG")
	fmt.Printf("DEBUG: %s\n", debug)

	cfg := &Config{}

	// Load environment variables into the Config struct using envconfig
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate the config struct
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Print out the loaded config (for testing purposes)
	log.Printf("Configuration Loaded: %+v\n\n", cfg)
	return cfg, nil
}

func NewTestConfig() (*Config, error) {
	cfg := &Config{
		Server: Server{
			Port: "8080",
		},
	}

	return cfg, nil
}
