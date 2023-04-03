package ws

import (
	"context"
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

	duration, err := time.ParseDuration(os.Getenv("SESSION_TIMEOUT"))
	if err != nil {
		duration = time.Minute
	}
	server.shutdownTimeout = duration

	return server
}

func (s *Server) authenticateRequest(r *http.Request) bool {
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

	// Check for HTTP basic auth using Authorization header.
	if !s.authenticateRequest(r) {
		log.Debug("Unable to authenticate - request rejected")

		w.WriteHeader(http.StatusUnauthorized)
		r.Close = true

		return
	}

	// Check session token for an already active session.
	// If tokens do not match - reject request.
	// The session will timeout on its own after Server.shutdownTimeout.
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

	// After the user has been granted access, stop session
	// shutdown timer if one exists.
	if s.shutdownTimer != nil {
		log.Debug("Stopping session shutdown timer")
		s.shutdownTimer.Stop()
		log.Info("Session shutdown aborted")
	}

	// If this is the first connection - initialize session context
	// to be passed down to the commands, otherwise do nothing.
	if !s.state.active {
		log.Info("Starting a new session")
		s.ctx.session.ctx, s.ctx.session.cancel = context.WithCancel(
			context.Background(),
		)
	} else {
		log.Info("Recovering previous session")
	}

	// Initialize new connection context that is dependant on session
	// context. Therefore, if session context is cancelled, connection
	// context is cancelled as well.
	s.ctx.connection.ctx, s.ctx.connection.cancel = context.WithCancel(
		s.ctx.session.ctx,
	)

	// Now that connection context has been initialized,
	// we can send the session token to the client and
	// initialize server state.
	if !s.state.active {
		s.state.initialize()
		s.updates = make(chan *Message)

		err = s.sendSessionToken(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Listen for updates and send them to the user until
	// either session or connection context is cancelled.
	go s.listenForUpdates(c)

	// Listen for command messages from the client.
	// If connection close request is encountered - it cancels session context.
	// If connection is lost - it cancels connection context leaving session
	// intact.
	s.listenForCommands(c)

	c.Close(websocket.StatusNormalClosure, "")

	log.Info("Connection closed")

	// If session context is not cancelled after client disconnect, then
	// start session shutdown timer which cancels session context
	if s.ctx.session.ctx.Err() == nil {
		log.Infof("Session is still active - shutting down in %s", s.shutdownTimeout)
		s.shutdownTimer = time.AfterFunc(s.shutdownTimeout, func() {
			s.closeSession()
			log.Info("Session shutdown successfully")
		})
	}
}

func (s *Server) closeSession() {
	if s.ctx.session.cancel != nil {
		s.ctx.session.cancel()
	}
	s.state.reset()
}

func (s *Server) closeConnection() {
	if s.ctx.connection.cancel != nil {
		s.ctx.connection.cancel()
	}
}

func (s *Server) sendSessionToken(c *websocket.Conn) error {
	msg, err := NewMessageWithBody("session", s.state.session)
	if err != nil {
		return err
	}

	msg.Type = Update

	return c.Write(s.ctx.connection.ctx, websocket.MessageText, msg.MustBytes())
}

func (s *Server) ConnectionEstablished() bool {
	return s.state.active
}

// Close currently running session and disconnect the client.
func (s *Server) Close() {
	s.closeSession()
}

// Drop client connection without closing currently running session.
// Mainly used for client reconnect testing.
func (s *Server) Drop() {
	s.closeConnection()
}
