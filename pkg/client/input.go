package robocat

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/robocat-ai/robocat/internal/ws"
)

type RobocatInput struct {
	client *Client
}

func (i *RobocatInput) Push(file *ws.RobocatFile) error {
	if i.client == nil {
		return errors.New("client is not set")
	}

	ref, err := i.client.sendCommand("input", file)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	i.client.subscribe(ref, func(ctx context.Context, m *ws.Message) {
		if m.Name != "status" {
			err = fmt.Errorf("unexpected update message: '%s'", m.Name)
		} else if m.MustText() != "ok" {
			err = fmt.Errorf("retured status was not 'ok': '%s'", m.MustText())
		} else if m.Ref != ref {
			err = errors.New("update message reference does not match the command")
		}

		wg.Done()
	})
	defer i.client.unsubscribe(ref)

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}
