package robocat

import (
	"bytes"
	"context"
	"fmt"
	_ "image/jpeg"
	"image/png"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlowCommand(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	flow := client.Flow("01-example-com").WithTimeout(15 * time.Second).Run()
	assert.NoError(t, flow.Err())

	flow.Log().Watch(func(line string) {
		// t.Log(line)
	})

	var outputError error

	flow.Files().Watch(func(file *File) {
		if file.Kind() == "text" {
			t.Logf("Client received TEXT: %s", file.Text())
			if file.Text() != "Example Domain" {
				outputError = fmt.Errorf("expected 'Example Domain', got '%s'", file.Text())
				return
			}

		} else if file.Kind() == "image" {
			buffer := bytes.NewBuffer(nil)
			image, format, err := file.Image()
			if err != nil {
				outputError = err
				return
			}

			err = png.Encode(buffer, image)
			if err != nil {
				outputError = err
				return
			}

			t.Logf("Client received IMAGE (%s): %s", format, file.Path)
		}
	})

	err := flow.Wait()
	assert.NoError(t, err)
	assert.NoError(t, outputError)
}

func TestMissingFlow(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	flow := client.Flow("missing-flow").Run()
	assert.NoError(t, flow.Err())

	err := flow.Wait()
	assert.ErrorContains(t, err, "cannot find missing-flow")
}

func TestFlowTimeout(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	flow := client.Flow("01-example-com").WithTimeout(time.Second).Run()
	assert.NoError(t, flow.Err())

	err := flow.Wait()
	assert.ErrorContains(t, err, context.DeadlineExceeded.Error())
}
