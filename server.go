package inspect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

type Server struct {
	request   *Response
	requestID int64
	logger    *zap.Logger
	err       error

	action chan requestAction

	handlerMutex sync.Mutex
	requestMutex sync.RWMutex
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

func (s *Server) start() (port int, err error) {
	port, err = findAvailablePort()
	if err != nil {
		return port, fmt.Errorf("cannot start caddy-inspect server: %w", err)
	}

	errChan := make(chan error)
	go func(server *Server) {
		errChan <- http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), server)
		s.handlerMutex.Lock()
		s.err = <-errChan
		s.handlerMutex.Unlock()
	}(s)

	return port, nil
}

func (s *Server) handle(ctx caddy.Context, w http.ResponseWriter, r *http.Request) requestAction {
	s.handlerMutex.Lock()
	defer s.handlerMutex.Unlock()

	request := buildResponse(ctx, w, r)

	s.requestMutex.Lock()
	s.request = &request
	s.requestID = time.Now().UnixNano()
	s.requestMutex.Unlock()

	action := <-s.action

	s.requestMutex.Lock()
	s.request = nil
	s.requestMutex.Unlock()

	return action
}

func writeJson(w http.ResponseWriter, body any) error {
	w.Header().Add("content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
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

	// all other endpoints
	w.Write(htmlContent)
}

var _ http.Handler = (*Server)(nil)
