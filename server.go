package inspect

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

type Server struct {
	// http handler
	request   *Response
	requestID int64
	action    chan requestAction

	// static handler
	static http.Handler

	// instance
	port   int
	logger *zap.Logger
	err    error

	instanceMutex sync.Mutex
	requestMutex  sync.RWMutex
}

type requestAction = int

const (
	requestActionResume requestAction = iota
	requestActionStep
	requestActionStop
)

func (s *Server) hasRequest() bool {
	s.requestMutex.RLock()
	defer s.requestMutex.RUnlock()

	return s.request != nil
}

func (s *Server) hasResponse() bool {
	s.requestMutex.RLock()
	defer s.requestMutex.RUnlock()

	return s.request != nil && s.request.responseMode
}

// start starts the server. If the server is already running, nothing is done.
//
// returns the port the server is listening on,
// whether the server has been previously started,
// and an error if any
func (s *Server) start() (port int, started bool, err error) {
	s.instanceMutex.Lock()
	defer s.instanceMutex.Unlock()

	// server already started
	if s.port > 0 {
		return s.port, true, nil
	}

	port, err = findAvailablePort()
	if err != nil {
		return port, false, fmt.Errorf("cannot start caddy-inspect server: %w", err)
	}
	s.port = port

	errChan := make(chan error)
	go func(server *Server) {
		errChan <- http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), server)
		s.instanceMutex.Lock()
		s.err = <-errChan
		s.instanceMutex.Unlock()
	}(s)

	return port, false, nil
}

func (s *Server) handle(m Middleware, w http.ResponseWriter, r *http.Request) requestAction {
	s.instanceMutex.Lock()
	defer s.instanceMutex.Unlock()

	request := buildResponse(m, w, r)

	s.requestMutex.Lock()
	s.request = &request
	s.requestID = time.Now().UnixNano()
	s.requestMutex.Unlock()

	action := <-s.action

	s.requestMutex.Lock()
	s.request = nil
	s.requestID = 0
	s.requestMutex.Unlock()

	return action
}

func writeJson(w http.ResponseWriter, body any) error {
	w.Header().Add("content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*") // this should probably be removed.
	return json.NewEncoder(w).Encode(body)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// post requests
	// resume and stop
	if r.Method == http.MethodPost {
		var action requestAction

		switch r.URL.Path {
		case "/stop":
			action = requestActionStop
		case "/step":
			action = requestActionStep
		default:
			action = requestActionResume
		}

		if s.hasRequest() {
			s.action <- action
		}

		if err := writeJson(w, "ok"); err != nil {
			s.logger.Error("error writing http response", zap.Error(err))
		}
		return
	}

	// handle /request
	if r.URL.Path == "/request" {
		var response struct {
			HasRequest  bool      `json:"has_request"`
			HasResponse bool      `json:"has_response"`
			Request     *Response `json:"request,omitempty"`
			ID          int64     `json:"id"`
		}

		response.HasRequest = s.hasRequest()
		response.HasResponse = s.hasResponse()
		response.Request = s.request
		response.ID = s.requestID

		if err := writeJson(w, response); err != nil {
			s.logger.Error("error writing http response", zap.Error(err))
		}
		return
	}

	s.static.ServeHTTP(w, r)
}

var _ http.Handler = (*Server)(nil)

var defaultServer *Server

func setUpServer(ctx caddy.Context) error {
	// setup already done
	if defaultServer != nil {
		return nil
	}

	dir, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("error setting up static file server: %w", err)
	}

	defaultServer = &Server{
		logger: ctx.Logger(),
		action: make(chan requestAction),
		static: http.FileServerFS(dir),
	}
	return nil
}

func getServerInstance() *Server { return defaultServer }
