package robocat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	assert.NoError(t, err)
}
