package robocat

import (
	"context"
	"log"
	"net/url"

	"nhooyr.io/websocket"
)

type Client struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	conn *websocket.Conn

	registeredCallbacks map[string][]UpdateCallback
}

func NewClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ctx:       ctx,
		ctxCancel: cancel,

		registeredCallbacks: make(map[string][]UpdateCallback),
	}

	return client
}

func (c *Client) Connect(u string, credentials ...Credentials) error {
	ur, err := url.Parse(u)
	if err != nil {
		return err
	}

	if len(credentials) > 0 {
		ur.User = credentials[0].GetUserInfo()
	}

	conn, _, err := websocket.Dial(
		c.ctx,
		ur.String(),
		&websocket.DialOptions{
			Subprotocols: []string{"robocat"},
		},
	)
	if err != nil {
		return err
	}

	c.conn = conn

	go c.listenForUpdates()

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.ctxCancel()
		return c.conn.Close(websocket.StatusNormalClosure, "")
	}

	return nil
}

func (c *Client) listenForUpdates() {
	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Stopped listening for updates")
			return
		default:
			msg, err := c.readUpdate()
			if err != nil {
				continue
			}

			c.broadcastEvent(c.ctx, msg)
		}
	}
}
