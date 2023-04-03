package robocat

import (
	"context"

	"github.com/robocat-ai/robocat/internal/ws"
)

type UpdateCallback func(context.Context, *ws.Message)

func (c *Client) subscribe(ref string, callback UpdateCallback) {
	_, ok := c.registeredCallbacks[ref]
	if !ok {
		c.registeredCallbacks[ref] = make([]UpdateCallback, 0)
	}

	c.registeredCallbacks[ref] = append(c.registeredCallbacks[ref], callback)
}

func (c *Client) unsubscribe(ref string) {
	c.registeredCallbacks[ref] = nil
}

func (c *Client) broadcastEvent(ctx context.Context, message *ws.Message) {
	c.logDebugf("<- recv: %s %s %s", message.Ref, message.Name, message.MustText())

	callbacks, ok := c.registeredCallbacks[message.Ref]
	if ok {
		for _, callback := range callbacks {
			go callback(ctx, message)
		}
	}
}
