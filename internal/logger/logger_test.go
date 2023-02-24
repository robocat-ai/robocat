package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMake(t *testing.T) {
	log := Make()
	log.Info("Test message")
}

func TestProductionLogger(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	log := Make()
	log.Info("Test message")
}
func TestMakeError(t *testing.T) {
	assert.PanicsWithError(t, "missing Level", func() {
		Make(zap.AtomicLevel{})
	})
}

func TestForModuleMake(t *testing.T) {
	log := ForModule("test")
	log.Info("Test message")
}
