package robocat

import (
	"errors"

	"github.com/oklog/ulid/v2"
	"github.com/robocat-ai/robocat/internal/ws"
	"nhooyr.io/websocket"
)

func newCommand(name string, body ...interface{}) (*ws.Message, error) {
	message, err := ws.NewMessageWithBody(name, body...)
	if err != nil {
		return nil, err
	}

	message.Type = ws.Command
	message.Ref = ulid.Make().String()

	return message, nil
}

func (c *Client) sendCommand(name string, body ...interface{}) (string, error) {
	message, err := newCommand(name, body...)
	if err != nil {
		return "", err
	}

	c.log("-> send:", message.Ref, message.Name, message.MustText())

	bytes, err := message.Bytes()
	if err != nil {
		return "", err
	}

	err = c.conn.Write(c.ctx, websocket.MessageText, bytes)
	if err != nil {
		return "", err
	}

	return message.Ref, nil
}

func updateFromBytes(bytes []byte) (*ws.Message, error) {
	message, err := ws.MessageFromBytes(bytes)
	if err != nil {
		return nil, err
	}

	if message.Type != ws.Update {
		return nil, errors.New("message is not an update")
	}

	return message, nil
}

func (c *Client) readUpdate() (*ws.Message, error) {
	_, bytes, err := c.conn.Read(c.ctx)
	if err != nil {
		return nil, err
	}

	msg, err := updateFromBytes(bytes)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
