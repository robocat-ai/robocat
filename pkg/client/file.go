package robocat

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"strings"
)

type File struct {
	Path     string
	MimeType string
	Payload  []byte
}

// Complete MIME-type as provided by mime.ParseMediaType.
func (f *File) Type() string {
	typ, _, err := mime.ParseMediaType(f.MimeType)
	if err != nil {
		return ""
	}

	return typ
}

// Generic file kind extracted from MIME-type (i.e. text, image, etc).
func (f *File) Kind() string {
	typ := f.Type()

	return strings.Trim(strings.Split(typ, "/")[0], "/")
}

// Converts transmitted payload to plain text and returns it as string.
func (f *File) Text() string {
	return string(f.Payload)
}

// Uses image.Decode to decode an image that has been encoded in a registered
// format. The string returned is the format name.
func (f *File) Image() (image.Image, string, error) {
	return image.Decode(bytes.NewReader(f.Payload))
}
