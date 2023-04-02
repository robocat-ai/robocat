package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

type ServerContext struct {
	session struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
	connection struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
}

type Server struct {
	Username string
	Password string

	state ServerState

	updates chan *Message

	registeredCallbacks map[string]CommandCallback

	ctx ServerContext

	shutdownTimeout time.Duration
	shutdownTimer   *time.Timer
}

func NewServer() *Server {
	server := &Server{
		registeredCallbacks: make(map[string]CommandCallback),
	}

	duration, err := time.ParseDuration("1s")
	if err != nil {
		duration = time.Second
	}
	server.shutdownTimeout = duration

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

	if !s.authenticateRequest(r) {
		log.Debug("Unable to authenticate - request rejected")

		w.WriteHeader(http.StatusUnauthorized)
		r.Close = true

		return
	}

	session := r.URL.Query().Get("session")
	if s.state.active && session != s.state.session {
		log.Debug("Invalid session token - request rejected")

		w.WriteHeader(http.StatusForbidden)
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

	if s.shutdownTimer != nil {
		log.Debug("Stopping session shutdown timer...")
		if !s.shutdownTimer.Stop() {
			<-s.shutdownTimer.C
		}
		log.Info("Session shutdown aborted")
	}

	if !s.state.active {
		log.Info("Starting a new session")
		s.ctx.session.ctx, s.ctx.session.cancel = context.WithCancel(context.Background())
	} else {
		log.Info("Recovering previous session")
	}

	s.ctx.connection.ctx, s.ctx.connection.cancel = context.WithCancel(s.ctx.session.ctx)

	if !s.state.active {
		s.state.initialize()
		s.updates = make(chan *Message)

		err = s.sendSessionToken(c, s.ctx.connection.ctx)
		if err != nil {
			log.Warn(err)
		}
	}

	go s.listenForUpdates(c)
	s.listenForCommands(c, s.ctx)

	c.Close(websocket.StatusNormalClosure, "")

	log.Info("Connection closed")

	s.shutdownTimeout, err = time.ParseDuration(os.Getenv("SESSION_TIMEOUT"))
	if err != nil {
		s.shutdownTimeout = time.Minute
	}

	if s.ctx.session.ctx.Err() == nil {
		log.Infof("Session is still active - shutting down in %s", s.shutdownTimeout)
		s.shutdownTimer = time.AfterFunc(s.shutdownTimeout, func() {
			s.ctx.session.cancel()
			s.state.reset()
			log.Info("Session shutdown successfully")
		})
	}
}

func (s *Server) sendSessionToken(
	c *websocket.Conn,
	ctx context.Context,
) error {
	msg, err := NewMessageWithBody("session", s.state.session)
	if err != nil {
		return err
	}

	msg.Type = Update

	return c.Write(ctx, websocket.MessageText, msg.MustBytes())
}

func (s *Server) ConnectionEstablished() bool {
	return s.state.active
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
	ctx ServerContext,
) {
	typ, bytes, err := c.Read(ctx.session.ctx)
	status := websocket.CloseStatus(err)

	if status != -1 {
		log.Debugw("Got close request", "status", status.String())
		log.Debug("Closing session")
		ctx.session.cancel()
		s.state.reset()
		return
	} else if err != nil {
		log.Debugw(
			"Got error while reading message",
			"error", err,
			"message", string(bytes),
		)
		log.Debug("Closing connection")
		ctx.connection.cancel()
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

		err = s.processCommand(ctx.session.ctx, command)
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

func (s *Server) listenForCommands(c *websocket.Conn, ctx ServerContext) {
	for {
		select {
		case <-ctx.session.ctx.Done():
		case <-ctx.connection.ctx.Done():
			return
		default:
			s.readCommand(c, ctx)
		}
	}
}

func (s *Server) listenForUpdates(c *websocket.Conn) {
	for {
		select {
		case <-s.ctx.session.ctx.Done():
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
