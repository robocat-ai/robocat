package ws

import (
	"encoding/json"
	"errors"
	"fmt"
)

type MessageType string

const (
	Update  MessageType = "update"
	Command MessageType = "command"
)

type Message struct {
	server *Server

	Type MessageType     `json:"type"`
	Name string          `json:"name"`
	Body json.RawMessage `json:"body,omitempty"`
	Ref  string          `json:"ref,omitempty"`
}

func (m *Message) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (m *Message) MustBytes() []byte {
	bytes, err := m.Bytes()
	if err != nil {
		return []byte{}
	}

	return bytes
}

func (m *Message) Text() (string, error) {
	var text *string
	err := json.Unmarshal(m.Body, &text)
	if err != nil {
		return "", err
	}

	return *text, nil
}

func (m *Message) MustText() string {
	text, err := m.Text()
	if err != nil {
		return ""
	}

	return text
}

func (m *Message) Reply(name string, body ...interface{}) error {
	if m.Type != Command {
		return errors.New("message is not a command")
	}

	if m.server == nil {
		return errors.New("message server must be set")
	}

	update, err := newUpdate(name, body...)
	if err != nil {
		return err
	}

	update.Ref = m.Ref
	return m.server.sendUpdate(update)
}

func (m *Message) ReplyWithError(err error) error {
	return m.Reply("error", err.Error())
}

func (m *Message) ReplyWithErrorf(format string, a ...any) error {
	return m.ReplyWithError(fmt.Errorf(format, a...))
}

func MessageFromBytes(bytes []byte) (*Message, error) {
	var message *Message

	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func NewMessageWithBody(name string, body ...interface{}) (*Message, error) {
	message := &Message{}

	message.Name = name

	var actualBody interface{} = nil

	if len(body) > 0 {
		actualBody = body[0]
	}

	if actualBody != nil {
		bytes, err := json.Marshal(actualBody)
		if err != nil {
			return nil, err
		}

		message.Body = json.RawMessage(bytes)
	}

	return message, nil
}

func newUpdate(name string, body ...interface{}) (*Message, error) {
	message, err := NewMessageWithBody(name, body...)
	if err != nil {
		return nil, err
	}

	message.Type = Update

	return message, nil
}

func commandFromBytes(bytes []byte) (*Message, error) {
	message, err := MessageFromBytes(bytes)
	if err != nil {
		return nil, err
	}

	if message.Type != Command {
		return nil, errors.New("message is not a command")
	}

	return message, nil
}
