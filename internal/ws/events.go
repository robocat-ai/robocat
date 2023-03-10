package ws

import "context"

type CommandCallback func(context.Context, *Message)

func (s *Server) On(name string, callback CommandCallback) {
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
