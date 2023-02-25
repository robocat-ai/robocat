package ws

type RobocatDataFields struct {
	Path     string `json:"path"`
	MimeType string `json:"type"`
	Payload  []byte `json:"payload"`
}
