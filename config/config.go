package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

type Logger struct {
	LogFile string `envconfig:"LOG_FILE" default:"./logs/server.log"`
}

type Server struct {
	Port string `envconfig:"SERVER_PORT" default:"8888" validate:"numeric"`
}

type Debug struct {
	Debug bool `envconfig:"DEBUG"`
}

type DB struct {
	Host            string `envconfig:"DB_HOST" default:"localhost"`
	Port            string `envconfig:"DB_PORT" default:"5432"`
	User            string `envconfig:"DB_USER" default:"postgres"`
	Password        string `envconfig:"DB_PASSWORD" default:"postgres"`
	Name            string `envconfig:"DB_NAME" default:"postgres"`
	File            string `envconfig:"DB_FILE" default:"./db/px.db"` // SQLite file path
	SSLMode         string `envconfig:"DB_SSL_MODE" default:"disable"`
	MaxIdleConns    int    `envconfig:"DB_MAX_IDLE_CONNS" default:"2"`
	MaxOpenConns    int    `envconfig:"DB_MAX_OPEN_CONNS" default:"5"`
	MaxConnLifetime int    `envconfig:"DB_MAX_CONN_LIFETIME" default:"10"`
}

type Swagger struct {
	FilePath string `envconfig:"SWAGGER_FILE_PATH" default:"./api/swagger/swagger.json"`
}

type CMD struct {
	Validate bool `envconfig:"CMD_VALIDATE" default:"false"`
}

type TaskLogger struct {
	DirPath string `envconfig:"TASK_LOGGER_DIR_PATH" default:"./task_logs"`
}

type Config struct {
	DB         DB
	Logger     Logger
	Server     Server
	Debug      Debug
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
