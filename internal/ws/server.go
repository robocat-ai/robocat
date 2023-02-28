package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"nhooyr.io/websocket"
)

type Server struct {
	Username string
	Password string

	active bool

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

func (s *Server) reset() {
	s.active = false
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := r.RemoteAddr

	log := log.With("client", client)

	log.Info("Got incoming connection")

	if s.active {
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

	close := &sync.WaitGroup{}
	close.Add(1)

	s.reset()

	s.active = true
	defer s.reset()

	go s.listenForCommands(close, c, ctx, client)
	go s.listenForUpdates(c, ctx)

	close.Wait()

	log.Info("Connection closed")
}

func (s *Server) ConnectionEstablished() bool {
	return s.active
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

func (s *Server) listenForCommands(
	close *sync.WaitGroup,
	c *websocket.Conn,
	ctx context.Context,
	client string,
) {
	s.commands = make(chan *Message)

	for {
		typ, bytes, err := c.Read(ctx)

		if err != nil {
			close.Done()
			return
		} else {
			if typ != websocket.MessageText {
				log.Debug("Only text messages are allowed")
				continue
			}

			log.With("command", string(bytes)).Debug("Received command")

			command, err := commandFromBytes(bytes)
			if err != nil {
				log.Debugf(
					"Got error while trying to parse command '%v': %s",
					string(bytes),
					err,
				)

				if !s.ConnectionEstablished() {
					close.Done()
					return
				}

				s.SendError(err)

				continue
			}

			err = s.processCommand(ctx, command)
			if err != nil {
				log.Debugf(
					"Got error while processing command '%s': %s",
					command.Name,
					err,
				)

				if !s.ConnectionEstablished() {
					close.Done()
					return
				}

				continue
			}
		}
	}
}

func (s *Server) listenForUpdates(c *websocket.Conn, ctx context.Context) {
	s.updates = make(chan *Message)

	for {
		update := <-s.updates

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
