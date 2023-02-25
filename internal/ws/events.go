package ws

import "context"

type RobocatCallback func(context.Context, *Message)

func (s *Server) On(name string, callback RobocatCallback) {
	s.registeredCallbacks[name] = callback
}

func (s *Server) broadcastEvent(
	ctx context.Context, name string, message *Message,
) {
	callback, ok := s.registeredCallbacks[name]
	if ok {
		go callback(ctx, message)
	}
}
