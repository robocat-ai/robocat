package ws

type RobocatDataFields struct {
	Path     string `json:"path"`
	MimeType string `json:"mime-type"`
	Payload  []byte `json:"payload"`
}
