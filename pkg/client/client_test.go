package robocat

import (
	"fmt"
	"testing"
	"time"
)

func newTestClient(t *testing.T) *Client {
	client, err := Connect(fmt.Sprintf("ws://%s", wsServerAddress), Credentials{
		wsServerUsername, wsServerPassword,
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
