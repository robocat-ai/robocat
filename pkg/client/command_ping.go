package robocat

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/robocat-ai/robocat/internal/ws"
)

func (c *Client) Ping() error {
	ref, err := c.sendCommand("ping")
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	c.subscribe(ref, func(ctx context.Context, m *ws.Message) {
		if m.Name != "pong" {
			err = fmt.Errorf("unexpected update message: '%s'", m.Name)
		} else if m.Ref != ref {
			err = errors.New("update message reference does not match the command")
		}

		wg.Done()
	})
	defer c.unsubscribe(ref)

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}
