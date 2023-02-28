package robocat

import (
	"errors"
	"fmt"
)

func (c *Client) Ping() error {
	ref, err := c.sendCommand("ping")
	if err != nil {
		return err
	}

	msg, err := c.readUpdate()
	if err != nil {
		return err
	}

	if msg.Name != "pong" {
		return fmt.Errorf("unexpected update message: '%s'", msg.Name)
	}

	if msg.Ref != ref {
		return errors.New("update message reference does not match the command")
	}

	return nil
}
