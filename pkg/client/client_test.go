package robocat

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/robocat-ai/robocat/internal/ws"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"
)

func newTestClient(t *testing.T) *Client {
    // Wait for session timeout to expire
	time.Sleep(4 * time.Second)

	client, err := Connect(fmt.Sprintf("ws://%s", wsServerAddress), ClientOptions{
		Credentials: Credentials{
			wsServerUsername, wsServerPassword,
		},
		ReconnectAttempts: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func TestClientConnect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	go func() {
		time.Sleep(time.Minute)
		client.Close()
	}()
}

func TestServerConnectionDrop(t *testing.T) {
	os.Setenv("SESSION_TIMEOUT", "3s")

	listener, err := nettest.NewLocalListener("tcp")
	require.NoError(t, err)

	address := fmt.Sprintf("ws://%s", listener.Addr().String())

	server := ws.NewServer()

	s := &http.Server{
		Handler:      server,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	var wg sync.WaitGroup

	counter := 1

	server.On("test", func(ctx context.Context, m *ws.Message) {
		t.Log("Running test command")

		for {
			select {
			case <-ctx.Done():
				t.Log("Context is cancelled")
				wg.Done()
				return
			default:
				t.Logf("Incrementing counter: %d", counter)
				m.Reply("count", counter)
				counter++

				if counter >= 10 {
					server.Close()
				} else if counter%5 == 0 {
					server.Drop()
				}

				time.Sleep(time.Second)
			}
		}
	})

	go func() {
		t.Logf("Listening on %v", address)
		s.Serve(listener)
		t.Logf("Stopped listening on %v", address)
	}()

	client, err := Connect(address, ClientOptions{
		ReconnectAttempts: 5,
	})
	require.NoError(t, err)

	setClientLogger(client, t)

	wg.Add(1)

	_, err = client.sendCommand("test")
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, 10, counter)
}

func TestServerReconnectExponentialBackoff(t *testing.T) {
	listener, err := nettest.NewLocalListener("tcp")
	require.NoError(t, err)

	address := fmt.Sprintf("ws://%s", listener.Addr().String())

	server := ws.NewServer()

	s := &http.Server{
		Handler:      server,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	var wg sync.WaitGroup

	server.On("connect", func(ctx context.Context, m *ws.Message) {
		t.Log("Shutting down the server")

		server.Close()

		wg.Done()
	})

	go func() {
		t.Logf("Listening on %v", address)
		s.Serve(listener)
		t.Logf("Stopped listening on %v", address)
	}()

	client, err := Connect(address, ClientOptions{
		ReconnectAttempts: 5,
	})
	require.NoError(t, err)

	setClientLogger(client, t)

	wg.Add(1)

	_, err = client.sendCommand("connect")
	require.NoError(t, err)

	wg.Wait()

	listener.Close()

	// Wait for 2 + 4 seconds (two attempts to reconnect)
	// + 1 more second just to be sure.
	time.Sleep(7 * time.Second)

	require.Equal(t, 8*time.Second, client.exponentialBackoffDelayDuration)
}
