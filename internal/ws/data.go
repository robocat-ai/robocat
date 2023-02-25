package ws

import (
	"encoding/json"
	"errors"
)

type RobocatDataFields struct {
	Path string `json:"path"`
	// Mime-Type of the file (set only for outputs to aid decoding the payload).
	MimeType string `json:"type"`
	Payload  []byte `json:"payload"`
}

func ParseDataFields(m *Message) (*RobocatDataFields, error) {
	if m.Name != "input" {
		return nil, errors.New("message is not of type 'input'")
	}

	var fields *RobocatDataFields
	err := json.Unmarshal(m.Body, &fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}
