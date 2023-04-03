package robocat

import (
	"context"
	"net/url"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

type Logger struct {
	Debugf func(format string, v ...any)
	Errorf func(format string, v ...any)
}

type ClientOptions struct {
	Credentials       Credentials
	ReconnectAttempts int // Default is zero, which means reconnects are disabled.
}

type ClientContext struct {
	session struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
	connection struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
}

type Client struct {
	ctx ClientContext

	logger *Logger

	conn *websocket.Conn
	err  error

	registeredCallbacks map[string][]UpdateCallback

	cancelFlow chan struct{}

	input *RobocatInput

	url *url.URL

	reconnectAttempts               int
	maxReconnectAttempts            int
	exponentialBackoffDelayDuration time.Duration
}

func makeClient(opts ClientOptions) *Client {

	client := &Client{
		registeredCallbacks:  make(map[string][]UpdateCallback),
		cancelFlow:           make(chan struct{}),
		maxReconnectAttempts: opts.ReconnectAttempts,
	}

	client.ctx.session.ctx, client.ctx.session.cancel = context.WithCancel(context.Background())

	client.resetExponentialBackoffDelay()

	return client
}

func Connect(u string, opts ...ClientOptions) (*Client, error) {
	var options ClientOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	client := makeClient(options)

	err := client.connect(u, options.Credentials)
	if err != nil {
		return nil, err
	}

	go client.listenForUpdates()

	return client, nil
}

func (c *Client) closeSession() {
	c.closeConnection()
	c.ctx.session.cancel()
}

func (c *Client) closeConnection() error {
	var err error
	if c.conn != nil {
		// We need to properly close the connection,
		// so the previous session gets closed properly.
		err = c.conn.Close(websocket.StatusNormalClosure, "")
	}

	if c.ctx.connection.cancel != nil {
		c.ctx.connection.cancel()
	}

	return err
}

func (c *Client) Close() error {
	return c.closeConnection()
}

func (c *Client) setCredentials(credentials Credentials) {
	c.url.User = credentials.GetUserInfo()
}

func (c *Client) connect(u string, credentials ...Credentials) (err error) {
	// Store connection URL to use later during automatic reconnects.
	c.url, err = url.Parse(u)
	if err != nil {
		return err
	}

	if len(credentials) > 0 {
		c.setCredentials(credentials[0])
	}

	// Close previously open connection, if any.
	c.Close()

	// Create new client context that is closed upon calling Client.Close().
	c.ctx.connection.ctx, c.ctx.connection.cancel = context.WithCancel(c.ctx.session.ctx)

	conn, _, err := websocket.Dial(
		c.ctx.connection.ctx,
		c.url.String(),
		&websocket.DialOptions{
			Subprotocols: []string{"robocat"},
		},
	)
	if err != nil {
		return err
	}

	// TODO: Set read limit to an incredibly high amount, since in
	// next version of websocket library the limit will be removed
	// by default: https://github.com/nhooyr/websocket/pull/256/commits/ea87744105d79f972e58404bb46791b97fc3f314
	const gigabyte int64 = 1024 * 1024 * 1024
	conn.SetReadLimit(gigabyte)

	c.conn = conn

	return nil
}

func (c *Client) logDebugf(format string, v ...any) {
	if c.logger != nil && c.logger.Debugf != nil {
		c.logger.Debugf(format, v...)
	}
}

func (c *Client) logErrorf(format string, v ...any) {
	if c.logger != nil && c.logger.Errorf != nil {
		c.logger.Errorf(format, v...)
	}
}

func (c *Client) SetLogger(logger *Logger) {
	c.logger = logger
}

func (c *Client) resetExponentialBackoffDelay() {
	c.exponentialBackoffDelayDuration = time.Second
}

func (c *Client) exponentialBackoffDelay() time.Duration {
	if c.exponentialBackoffDelayDuration <= time.Minute {
		c.exponentialBackoffDelayDuration = 2 * c.exponentialBackoffDelayDuration
	}

	return c.exponentialBackoffDelayDuration
}

func (c *Client) listenForUpdates() {
	for {
		select {
		case <-c.ctx.session.ctx.Done():
			c.logDebugf("Session context cancelled - stopping listening")
			return
		case <-c.ctx.connection.ctx.Done():
			c.logDebugf("Connection context cancelled - stopping listening")
			return
		default:
			message, err := c.readUpdate()
			status := websocket.CloseStatus(err)

			if status != -1 {
				c.logDebugf("WebWocket is closed (%s) - closing connection", status.String())
				c.closeSession()
				return
			} else if err != nil {
				c.logErrorf("Got error while reading message: %v", err)
				delay := c.exponentialBackoffDelay()
				if strings.Contains(err.Error(), "WebSocket closed") {
					c.logDebugf("WebSocket closed - closing connection")
					c.closeSession()
				} else if c.maxReconnectAttempts == 0 {
					c.logDebugf("No reconnects - closing connection")
					c.closeSession()
					return
				} else if c.reconnectAttempts >= c.maxReconnectAttempts {
					c.logDebugf("Maximum number of reconnects reached - closing connection")
					c.closeSession()
					return
				}

				c.logDebugf("Trying to reconnect in %s...", delay)
				time.Sleep(delay)
				c.reconnectAttempts++

				c.err = c.connect(c.url.String())
				if c.err != nil {
					c.logErrorf("Got error while trying to reconnect: %v", c.err)
				}

				continue
			} else if message.Name == "session" {
				values := c.url.Query()
				values.Set("session", message.MustText())
				c.url.RawQuery = values.Encode()
				continue
			}

			c.resetExponentialBackoffDelay()
			c.broadcastEvent(c.ctx.connection.ctx, message)
		}
	}
}

// cancelFlowChannel returns a channel that's closed
// when child flow of this client should be canceled.
func (c *Client) cancelFlowChannel() <-chan struct{} {
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
