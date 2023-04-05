package ws

import "github.com/oklog/ulid/v2"

type ServerState struct {
	active  bool
	session string
}

func (s *ServerState) reset() {
	*s = ServerState{}
}

func (s *ServerState) initialize() {
	s.reset()
	s.active = true
	s.session = ulid.Make().String()
}
