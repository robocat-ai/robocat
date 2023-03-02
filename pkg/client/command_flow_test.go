package robocat

import (
	"log"
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

	defer flow.Close()

	logger := flow.Log()

	go func() {
		for {
			line, err := logger.Next()
			if err != nil {
				continue
			}

			log.Println(line)
		}
	}()

	err := flow.Wait()
	assert.NoError(t, err)
}
