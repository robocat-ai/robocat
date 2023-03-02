package robocat

import (
	"context"
	"log"
	"net/url"

	"github.com/docker/go-units"
	"github.com/sakirsensoy/genv"
	"nhooyr.io/websocket"
)

type Client struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	conn *websocket.Conn
	err  error

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
	url, err := url.Parse(u)
	if err != nil {
		return err
	}

	if len(credentials) > 0 {
		url.User = credentials[0].GetUserInfo()
	}

	conn, _, err := websocket.Dial(
		c.ctx,
		url.String(),
		&websocket.DialOptions{
			Subprotocols: []string{"robocat"},
		},
	)
	if err != nil {
		return err
	}

	c.conn = conn

	size, err := units.FromHumanSize(
		genv.Key("MAX_READ_SIZE").Default("1M").String(),
	)
	if err != nil {
		log.Fatal(err)
	}

	c.conn.SetReadLimit(size)

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
			return
		default:
			message, err := c.readUpdate()
			if err != nil {
				c.err = err
				c.Close()
				continue
			}

			c.broadcastEvent(c.ctx, message)
		}
	}
}
