package robocat

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRobocatStream(t *testing.T) {
	items := 3

	wg := sync.WaitGroup{}
	wg.Add(2 * items)

	stream := &RobocatStream[string]{}
	defer stream.Close()

	go func() {
		for {
			item, ok := <-stream.Channel()
			if !ok {
				break
			}

			t.Log(item)
			wg.Done()
		}
	}()

	go func() {
		for i := 1; i <= items; i++ {
			stream.put(fmt.Sprintf("item %d", i))
			wg.Done()
		}
	}()

	wg.Wait()
}

func TestStreamClose(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	stream := &RobocatStream[string]{}

	stream.Close()

	var err error

	go func() {
		err = stream.put("")
		wg.Done()
	}()

	wg.Wait()

	assert.ErrorContains(t, err, "stream channel is closed")
}
