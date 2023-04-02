package ws

type ServerState struct {
	active bool
}

func (s *ServerState) reset() {
	*s = ServerState{}
}

func (s *ServerState) initialize() {
	s.reset()
	s.active = true
}
