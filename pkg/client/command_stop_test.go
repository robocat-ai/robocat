package robocat

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopCommand(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip()
	}

	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	flow := client.Flow("02-long-polling").WithTimeout(15 * time.Second).Run()
	assert.NoError(t, flow.Err())

	time.Sleep(5 * time.Second)

	err := client.Stop()
	assert.NoError(t, err)

	err = flow.Wait()
	assert.ErrorContains(t, err, "flow was aborted")
}
