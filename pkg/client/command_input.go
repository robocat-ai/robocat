package robocat

import (
	"mime"
	"path/filepath"
)

func (c *Client) Input(path string, content []byte) error {
	mimeType := mime.TypeByExtension(filepath.Ext(path))

	return c.getInput().Push(path, mimeType, content)
}
