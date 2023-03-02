package robocat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlowCommand(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// client.DebugLogger(log.Println)

	flow := client.Flow("01-example-com").WithTimeout(15 * time.Second).Run()
	assert.NoError(t, flow.Err())

	go func() {
		log := flow.Log()

		for {
			line, ok := <-log.Channel()
			if !ok {
				break
			}

			t.Log(line)
		}
	}()

	err := flow.Wait()
	assert.NoError(t, err)
}

func TestMissingFlow(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	flow := client.Flow("missing-flow").Run()
	assert.NoError(t, flow.Err())

	err := flow.Wait()
	assert.ErrorContains(t, err, "cannot find missing-flow")
}
