package logger

import (
	"fmt"
	"os"

	"github.com/fattymango/px-take-home/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger struct with SugaredLogger
type Logger struct {
	*zap.SugaredLogger
	config *config.Config
}

// NewLogger returns a new production logger backed by zap
func NewLogger(cfg *config.Config) (*Logger, error) {
	fmt.Println("Path to log file:", cfg.Logger.LogFile)
	fmt.Println("Debug mode:", cfg.Debug.Debug)

	// Zap production config
	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if cfg.Debug.Debug {
		conf.Level.SetLevel(zap.DebugLevel)
	}

	conf.EncoderConfig.TimeKey = "timestamp"
	conf.EncoderConfig.CallerKey = "caller"
	conf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")

	// File output setup via lumberjack
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Logger.LogFile,
		MaxSize:    1,  // megabytes
		MaxBackups: 30, // number of backups
		MaxAge:     30, // days
	})

	// Console output
	consoleWriter := zapcore.Lock(os.Stdout)

	// File encoder for JSON logging
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	// Console encoder for human-readable output
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00"),
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})

	// Create cores for file and console logging
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, conf.Level)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, conf.Level)

	// Combine file and console cores
	combinedCore := zapcore.NewTee(fileCore, consoleCore)

	// Build the logger with the combined core
	zapLogger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	// Return the SugaredLogger
	return &Logger{
		zapLogger.Sugar(),
		cfg,
	}, nil
}

// NewTestLogger returns a logger configured for testing with development settings
func NewTestLogger() *Logger {
	zapLogger := zap.NewNop().Sugar()
	return &Logger{
		SugaredLogger: zapLogger,
		config:        &config.Config{},
	}
}
