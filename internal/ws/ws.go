package ws

import (
	"net"
	"net/http"
	"time"

	"github.com/robocat-ai/robocat/internal/shared"
	"github.com/sakirsensoy/genv"
)

func Start(options shared.Options) {
	log.Info("Starting WebSocket module...")

	username := genv.Key("AUTH_USERNAME").String()
	password := genv.Key("AUTH_PASSWORD").String()

	if len(username) == 0 && len(password) == 0 {
		log.Warn("No AUTH_USERNAME and AUTH_PASSWORD specified - anybody can connect!")
	}

	listener, err := net.Listen("tcp", options.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()

	server.Username = username
	server.Password = password

	s := &http.Server{
		Handler:      server,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	runner := NewRobocatRunner()
	server.On("run", runner.Handle)
	server.On("stop", runner.Stop)
	server.On("input", runner.GetInput().Handle)

	log.Infof("Listening on ws://%v", listener.Addr())

	err = s.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
