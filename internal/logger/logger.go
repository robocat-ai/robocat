package logger

import (
	"log"

	"github.com/robocat-ai/robocat/internal/utils"
	"go.uber.org/zap"
)

func getConfig() zap.Config {
	var config zap.Config

	if utils.IsLocal() {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	if !utils.IsProduction() {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return config
}

// Logger factory function. Returns a new instance of zap sugared logger.
func Make(level ...zap.AtomicLevel) *zap.SugaredLogger {
	c := getConfig()

	if len(level) > 0 {
		c.Level = level[0]
	}

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
