package robocat

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientConnect(t *testing.T) {
	client := NewClient()

	go func() {
		time.Sleep(time.Hour)
		client.Close()
	}()

	err := client.Connect(fmt.Sprintf("ws://%s", wsServerAddress), Credentials{
		wsServerUsername, wsServerPassword,
	})
	assert.NoError(t, err)

	defer client.Close()
}
