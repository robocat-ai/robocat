package robocat

import (
	"context"
	"net/url"

	"nhooyr.io/websocket"
)

type Client struct {
	ctx  context.Context
	conn *websocket.Conn

	// updates chan *Message
	// commands chan *Message
}

func NewClient() *Client {
	client := &Client{
		ctx: context.Background(),
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

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "")
	}

	return nil
}
