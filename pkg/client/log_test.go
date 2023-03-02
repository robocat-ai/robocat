package robocat

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	logLines := 3

	wg := sync.WaitGroup{}
	wg.Add(2 * logLines)

	log := &RobocatLog{}
	defer log.Close()

	go func() {
		for {
			line, ok := <-log.Channel()
			if !ok {
				break
			}

			t.Log(line)
			wg.Done()
		}
	}()

	go func() {
		for i := 1; i <= logLines; i++ {
			log.append(fmt.Sprintf("log line %d", i))
			wg.Done()
		}
	}()

	wg.Wait()
}

func TestLogClose(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	log := &RobocatLog{}

	log.Close()

	var err error

	go func() {
		err = log.append("")
		wg.Done()
	}()

	wg.Wait()

	assert.ErrorContains(t, err, "log channel is closed")
}
