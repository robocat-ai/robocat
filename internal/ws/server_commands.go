package ws

import (
	"fmt"

	"nhooyr.io/websocket"
)

func (s *Server) listenForCommands(c *websocket.Conn) {
	for {
		select {
		case <-s.ctx.session.ctx.Done():
			return
		case <-s.ctx.connection.ctx.Done():
			return
		default:
			s.readCommand(c)
		}
	}
}

func (s *Server) readCommand(c *websocket.Conn) {
	typ, bytes, err := c.Read(s.ctx.session.ctx)
	status := websocket.CloseStatus(err)

	if status != -1 {
		// Cancel session context if client sends close request.
		log.Debugw("Got close request", "status", status.String())
		log.Debug("Closing session")
		s.closeSession()
		return
	} else if err != nil {
		// Cancel connection context if client loses connection,
		// leaving session context active and giving the client
		// ability to reconnect and continue work.
		log.Debugw(
			"Got error while reading message",
			"error", err,
			"message", string(bytes),
		)
		log.Debug("Closing connection")
		s.closeConnection()
		return
	} else {
		if typ != websocket.MessageText {
			log.Debug("Only text messages are allowed")
			s.SendError(fmt.Errorf("only text messages are allowed"))
			return
		}

		log.With("message", string(bytes)).Debug("Received message")

		command, err := commandFromBytes(bytes)
		if err != nil {
			log.Debugw(
				"Got error while trying to parse command",
				"command", string(bytes),
				"error", err,
			)

			s.SendError(err)

			return
		}

		err = s.processCommand(command)
		if err != nil {
			log.Debugw(
				"Got error while processing command",
				"command", command.Name,
				"error", err,
			)

			s.SendError(err)

			return
		}
	}
}

func (s *Server) processCommand(message *Message) error {
	message.server = s

	if message.Name == "ping" {
		return message.Reply("pong")
	} else {
		s.broadcastEvent(s.ctx.session.ctx, message.Name, message)
	}

	return nil
}
