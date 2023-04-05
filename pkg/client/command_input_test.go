package robocat

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

type InputTestFile struct {
	Path    string
	Payload []byte
}

func TestInput(t *testing.T) {
	t.Skip() // TODO: Investigate why this test hangs

	client := newTestClient(t)
	defer client.Close()

	setClientLogger(client, t)

	files := []*InputTestFile{
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

		err = client.Input(file.Path, file.Payload)
		assert.NoError(t, err)

		// bytes, err = os.ReadFile(path.Join(targetDir, file.Path))
		// assert.NoError(t, err)
		// assert.Equal(t, file.Payload, bytes)
	}
}
