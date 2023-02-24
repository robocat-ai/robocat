package logger

import (
	"log"

	"github.com/robocat-ai/robocat/internal/utils"
	"github.com/sakirsensoy/genv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getConfig() zap.Config {
	var config zap.Config

	if utils.IsLocal() {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	level, err := zap.ParseAtomicLevel(
		genv.Key("LOG_LEVEL").Default("info").String(),
	)
	if err == nil {
		config.Level = level
	}

	return config
}

// Logger factory function. Returns a new instance of zap sugared logger.
func Make() *zap.SugaredLogger {
	c := getConfig()

	l, err := c.Build()

	if err != nil {
		log.Printf("Got error during logger initialization: %s", err)
		panic(err)
	}

	return l.Sugar()
}

func ForModule(module string) *zap.SugaredLogger {
	return Make().Named("module").Named(module)
}
