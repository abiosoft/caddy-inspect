package inspect

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	request   *Request
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
	requestActionStop
)

func (s *Server) hasRequest() bool {
	s.requestMutex.RLock()
	defer s.requestMutex.RUnlock()

	return s.request != nil
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

func (s *Server) handle(r Request) requestAction {
	s.handlerMutex.Lock()
	defer s.handlerMutex.Unlock()

	s.requestMutex.Lock()
	s.request = &r
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
		action := requestActionResume
		if r.URL.Path == "/stop" {
			action = requestActionStop
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
			HasRequest bool     `json:"has_request"`
			Request    *Request `json:"request,omitempty"`
			ID         int64    `json:"id"`
		}

		response.HasRequest = s.hasRequest()
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

const httpServerListenPort = 2020

// findAvailablePort returns an available port on the host machine.
// it attempts port 2020 up till 2029
func findAvailablePort() (int, error) {
	allocatePort := func(port int) error {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return fmt.Errorf("error picking an available port: %w", err)
		}

		if err := listener.Close(); err != nil {
			return fmt.Errorf("error closing temporary port listener: %w", err)
		}

		return nil
	}

	var err error
	for i := 0; i < 10; i++ {
		port := httpServerListenPort + i
		err = allocatePort(port)
		if err == nil {
			return port, nil
		}
	}

	return 0, err
}
