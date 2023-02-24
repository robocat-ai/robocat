package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robocat-ai/robocat/internal/shared"
	"github.com/robocat-ai/robocat/internal/ws"
)

func main() {
	godotenv.Load()
	rand.Seed(time.Now().UnixMilli())
	options := shared.InitializeOptions()

	log.Infof("Starting robocat...")

	go ws.Start(options)

	waitForExitSignal()
}

func waitForExitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
}
