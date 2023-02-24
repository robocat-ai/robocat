package ws

import (
	"encoding/json"
	"errors"
)

type MessageType string

const (
	Update  MessageType = "update"
	Command MessageType = "command"
)

type Message struct {
	Type MessageType     `json:"type"`
	Name string          `json:"name"`
	Body json.RawMessage `json:"body,omitempty"`
}

func (m *Message) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func MessageFromBytes(bytes []byte) (*Message, error) {
	var message *Message

	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func NewUpdateWithBody(name string, body ...interface{}) (*Message, error) {
	message := &Message{
		Type: "update",
		Name: name,
	}

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

func NewUpdate(name string) *Message {
	message, _ := NewUpdateWithBody(name)

	return message
}

func CommandFromBytes(bytes []byte) (*Message, error) {
	message, err := MessageFromBytes(bytes)
	if err != nil {
		return nil, err
	}

	if message.Type != "command" {
		return nil, errors.New("message is not a command")
	}

	return message, nil
}
