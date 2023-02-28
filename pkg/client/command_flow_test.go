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

	flow := client.Flow("test").WithTimeout(5 * time.Second).Run()
	assert.NoError(t, flow.Err())

	defer flow.Close()

	flow.Wait()

	log.Println(flow.Err())
}
