package robocat

import (
	"github.com/robocat-ai/robocat/internal/ws"
)

func (c *Client) Input(file *ws.RobocatFile) error {
	return c.getInput().Push(file)
}
