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

	logger func(args ...any)

	conn *websocket.Conn
	err  error

	registeredCallbacks map[string][]UpdateCallback

	cancelFlow chan struct{}

	input *RobocatInput
}

func makeClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ctx:       ctx,
		ctxCancel: cancel,

		registeredCallbacks: make(map[string][]UpdateCallback),

		cancelFlow: make(chan struct{}),
	}

	return client
}

func Connect(u string, credentials ...Credentials) (*Client, error) {
	client := makeClient()

	url, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if len(credentials) > 0 {
		url.User = credentials[0].GetUserInfo()
	}

	conn, _, err := websocket.Dial(
		client.ctx,
		url.String(),
		&websocket.DialOptions{
			Subprotocols: []string{"robocat"},
		},
	)
	if err != nil {
		return nil, err
	}

	client.conn = conn

	size, err := units.FromHumanSize(
		genv.Key("MAX_READ_SIZE").Default("1M").String(),
	)
	if err != nil {
		log.Fatal(err)
	}

	client.conn.SetReadLimit(size)

	go client.listenForUpdates()

	return client, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.ctxCancel()
		return c.conn.Close(websocket.StatusNormalClosure, "")
	}

	return nil
}

func (c *Client) log(args ...any) {
	if c.logger != nil {
		c.logger(args...)
	}
}

func (c *Client) DebugLogger(logger func(args ...any)) {
	c.logger = logger
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

// CancelFlow returns a channel that's closed
// when child flow of this client should be canceled.
func (c *Client) CancelFlow() <-chan struct{} {
	return c.cancelFlow
}

func (c *Client) getInput() *RobocatInput {
	if c.input == nil {
		c.input = &RobocatInput{
			client: c,
		}
	}

	return c.input
}
