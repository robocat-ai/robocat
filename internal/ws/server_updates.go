package ws

import (
	"errors"
	"fmt"

	"nhooyr.io/websocket"
)

func (s *Server) listenForUpdates(c *websocket.Conn) {
	for {
		select {
		case <-s.ctx.session.ctx.Done():
			return
		case <-s.ctx.connection.ctx.Done():
			return
		case update := <-s.updates:
			if !s.ConnectionEstablished() {
				log.Debug("Handshake was not established yet")
				continue
			}

			bytes, err := update.Bytes()
			if err != nil {
				log.Debugf(
					"Got error while trying to send update '%v': %s",
					string(bytes), err,
				)
				continue
			}

			err = c.Write(s.ctx.connection.ctx, websocket.MessageText, bytes)
			if err != nil {
				log.Debugf(
					"Got error while trying to send update '%v': %s",
					string(bytes), err,
				)
				continue
			}
		}
	}
}

func (s *Server) sendUpdate(update *Message) error {
	if !s.ConnectionEstablished() {
		err := errors.New("connection was not established yet")
		return err
	}

	if s.updates != nil {
		select {
		case <-s.ctx.session.ctx.Done():
			return s.ctx.session.ctx.Err()
		case s.updates <- update:
			return nil
		}
	}

	return errors.New("update channel has not been initialized")
}

func (s *Server) Send(name string, body ...interface{}) error {
	update, err := newUpdate(name, body...)
	if err != nil {
		return err
	}

	return s.sendUpdate(update)
}

func (s *Server) SendError(err error) error {
	return s.Send("error", err.Error())
}

func (s *Server) SendErrorf(format string, a ...any) error {
	return s.SendError(fmt.Errorf(format, a...))
}
