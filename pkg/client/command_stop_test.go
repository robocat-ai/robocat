package robocat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopCommand(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	flow := client.Flow("02-stop-test").WithTimeout(15 * time.Second).Run()
	assert.NoError(t, flow.Err())

	time.Sleep(5 * time.Second)

	err := client.Stop()
	assert.NoError(t, err)

	flow.Wait()

	assert.ErrorContains(t, flow.Err(), "flow was aborted")
}
