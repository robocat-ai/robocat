package robocat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlowCommand(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// client.DebugLogger(t.Log)

	flow := client.Flow("01-example-com").WithTimeout(15 * time.Second).Run()
	assert.NoError(t, flow.Err())

	go flow.Log().Watch(func(line string) {
		t.Log(line)
	})

	err := flow.Wait()
	assert.NoError(t, err)
}

func TestMissingFlow(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// client.DebugLogger(t.Log)

	flow := client.Flow("missing-flow").Run()
	assert.NoError(t, flow.Err())

	err := flow.Wait()
	assert.ErrorContains(t, err, "cannot find missing-flow")
}
