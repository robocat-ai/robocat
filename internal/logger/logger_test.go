package logger

import (
	"os"
	"testing"
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

func TestForModule(t *testing.T) {
	log := ForModule("test")
	log.Info("Test message")
}
