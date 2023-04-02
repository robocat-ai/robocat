package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"nhooyr.io/websocket"
)

type Server struct {
	Username string
	Password string

	state ServerState

	updates  chan *Message
	commands chan *Message

	registeredCallbacks map[string]CommandCallback
}

func NewServer() *Server {
	server := &Server{
		registeredCallbacks: make(map[string]CommandCallback),
	}

	return server
}

func (s *Server) authenticateRequest(r *http.Request) bool {
	// return len(s.apiKey) > 0 && r.URL.Query().Get("key") != s.apiKey

	if len(s.Username) > 0 || len(s.Password) > 0 {
		username, password, ok := r.BasicAuth()
		if !ok {
			return false
		}

		return s.Username == username && s.Password == password
	}

	return true
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := r.RemoteAddr

	log := log.With("client", client)

	log.Info("Got incoming connection")

	if s.state.active {
		log.Debug("There is already an active connection - request rejected")

		w.WriteHeader(http.StatusBadRequest)
		r.Close = true

		return
	}

	if !s.authenticateRequest(r) {
		log.Debug("Unable to authenticate - request rejected")

		w.WriteHeader(http.StatusUnauthorized)
		r.Close = true

		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"robocat"},
	})
	if err != nil {
		log.Error(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "server closed connection")

	if c.Subprotocol() != "robocat" {
		log.Info("Connection closed - client must speak the robocat subprotocol")
		c.Close(websocket.StatusPolicyViolation, "client must speak the robocat subprotocol")
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	s.state.initialize()
	defer s.state.reset()

	go s.listenForCommands(c, ctx, cancel)
	go s.listenForUpdates(c, ctx)

	<-ctx.Done()

	c.Close(websocket.StatusNormalClosure, "")

	log.Info("Connection closed")
}

func (s *Server) ConnectionEstablished() bool {
	return s.state.active
}

func (s *Server) sendUpdate(update *Message) error {
	if !s.ConnectionEstablished() {
		err := errors.New("connection was not established yet")
		return err
	}

	s.updates <- update

	return nil
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

func (s *Server) processCommand(ctx context.Context, message *Message) error {
	message.server = s

	if message.Name == "ping" {
		return message.Reply("pong")
	} else {
		s.broadcastEvent(ctx, message.Name, message)
	}

	return nil
}

func (s *Server) readCommand(
	c *websocket.Conn,
	ctx context.Context,
	cancel context.CancelFunc,
) {
	typ, bytes, err := c.Read(ctx)
	status := websocket.CloseStatus(err)

	if status != -1 {
		log.Debugw("Got close request", "status", status.String())
		cancel()
		return
	} else if !s.ConnectionEstablished() {
		log.Debug("Read message while connection was not established")
		cancel()
		return
	} else if err != nil {
		log.Debugw(
			"Got error while reading message",
			"error", err,
			"message", string(bytes),
		)
		cancel()
		return
	} else {
		if typ != websocket.MessageText {
			log.Debug("Only text messages are allowed")
			s.SendError(fmt.Errorf("only text messages are allowed"))
			return
		}

		log.With("command", string(bytes)).Debug("Received command")

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

		err = s.processCommand(ctx, command)
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

func (s *Server) listenForCommands(
	c *websocket.Conn,
	ctx context.Context,
	cancel context.CancelFunc,
) {
	s.commands = make(chan *Message)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.readCommand(c, ctx, cancel)
		}
	}
}

func (s *Server) listenForUpdates(c *websocket.Conn, ctx context.Context) {
	s.updates = make(chan *Message)

	for {
		select {
		case <-ctx.Done():
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

			err = c.Write(ctx, websocket.MessageText, bytes)
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
