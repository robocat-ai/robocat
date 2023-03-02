package ws

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"strings"
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

// Complete MIME-type as provided by mime.ParseMediaType.
func (f *RobocatFile) Type() string {
	typ, _, err := mime.ParseMediaType(f.MimeType)
	if err != nil {
		return ""
	}

	return typ
}

// Generic file kind extracted from MIME-type (i.e. text, image, etc).
func (f *RobocatFile) Kind() string {
	typ := f.Type()

	return strings.Trim(strings.Split(typ, "/")[0], "/")
}

// Converts transmitted payload to plain text and returns it as string.
func (f *RobocatFile) Text() string {
	return string(f.Payload)
}

// Uses image.Decode to decode an image that has been encoded in a registered
// format. The string returned is the format name.
func (f *RobocatFile) Image() (image.Image, string, error) {
	return image.Decode(bytes.NewReader(f.Payload))
}
