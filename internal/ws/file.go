package ws

import (
	"encoding/json"
)

type RobocatFile struct {
	Path string `json:"path"`
	// Mime-Type of the file (set only for outputs to aid decoding the payload).
	MimeType string `json:"type"`
	Payload  []byte `json:"payload"`
}

func ParseFileFromMessage(m *Message) (*RobocatFile, error) {
	var fields *RobocatFile
	err := json.Unmarshal(m.Body, &fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}
