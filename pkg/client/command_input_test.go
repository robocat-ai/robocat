package robocat

import (
	"os"
	"path"
	"testing"

	"github.com/robocat-ai/robocat/internal/ws"
	"github.com/stretchr/testify/assert"
)

func TestInput(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	client.DebugLogger(t.Log)

	files := []*ws.RobocatFile{
		{
			Path: "sample-text",
		},
		{
			Path: "images/gopher.png",
		},
		{
			Path: "images/gopher.jpg",
		},
	}

	sourceDir := "test-input"
	targetDir := "test-flow/input"

	err := os.RemoveAll(targetDir)
	assert.NoError(t, err)

	for _, file := range files {
		bytes, err := os.ReadFile(path.Join(sourceDir, file.Path))
		assert.NoError(t, err)

		file.Payload = bytes

		err = client.Input(file)
		assert.NoError(t, err)

		bytes, err = os.ReadFile(path.Join(targetDir, file.Path))
		assert.NoError(t, err)
		assert.Equal(t, file.Payload, bytes)
	}
}
