package robocat

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/robocat-ai/robocat/internal/ws"
)

func (c *Client) Stop() error {
	ref, err := c.sendCommand("stop")
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	c.subscribe(ref, func(ctx context.Context, m *ws.Message) {
		if m.Name != "status" {
			err = fmt.Errorf("unexpected update message: '%s'", m.Name)
		} else if m.MustText() != "ok" {
			err = fmt.Errorf("retured status was not 'ok': '%s'", m.MustText())
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

	// c.ctxCancel()

	c.cancelFlow <- struct{}{}

	return nil
}
