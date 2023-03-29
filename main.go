package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
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

	if options.ProfilerEnabled {
		go func() {
			log.Debug("Running profiler server on port 6060")
			log.Debug("Usage: https://pkg.go.dev/net/http/pprof#hdr-Usage_examples")
			log.Debug(http.ListenAndServe(":6060", nil))
		}()
	}

	waitForExitSignal()
}

func waitForExitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
}
