package logger

import (
	"log"

	"github.com/fredytarigan/Tendo/pkg/tendo/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func getLoggerLevel(level string) zapcore.Level {
	if level == "debug" {
		return zap.DebugLevel
	}

	return zap.InfoLevel
}

func init() {
	var err error
	var logLevel string
	var is_development bool

	config.SetConfigFile("./config")

	if config.LoadConfig().AppMode != "Production" {
		logLevel = "debug"
		is_development = true
	} else {
		logLevel = "info"
		is_development = false
	}

	level := zap.NewAtomicLevelAt(getLoggerLevel(logLevel))
	encoder := zap.NewProductionEncoderConfig()

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig = encoder
	zapConfig.Level = level
	zapConfig.Development = is_development
	zapConfig.Encoding = "json"
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.OutputPaths = []string{"stdout"}
	
	logger, err := zapConfig.Build()

	if err != nil {
		log.Fatal("unable to construct logger with error: ", err)
	}

	Logger = logger
}
