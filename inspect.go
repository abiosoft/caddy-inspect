package inspect

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("inspect", parseCaddyfile)

}

var errRequestTerminated = errors.New("request terminated")

// Middleware implements an HTTP handler that writes the
// inspects the current request.
type Middleware struct {
	logger *zap.Logger
	ctx    caddy.Context
	server *Server
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.inspect",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger()
	m.ctx = ctx

	m.server = &Server{logger: m.logger, action: make(chan requestAction)}

	port, err := m.server.start()
	if err != nil {
		return fmt.Errorf("error occured during provision: %w", err)
	}

	m.logger.Info(fmt.Sprintf("inspect console available at http://127.0.0.1:%d", port))
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.logger.Debug("request is being inspected")

	action := m.server.handle(m.ctx, nil, r)

	switch action {
	case requestActionResume:
		m.logger.Debug("request resumed")
		return next.ServeHTTP(w, r)

	case requestActionStep:
		m.logger.Debug("request proceeding to response")

		// process middleware chain
		err := next.ServeHTTP(w, r)

		// handle the updated request details
		action := m.server.handle(m.ctx, w, r)
		if err != nil {
			return err
		}
		if action == requestActionResume {
			return nil
		}

		// request stopped
		fallthrough
	case requestActionStop:
		m.logger.Debug("request stopped")
	}

	return errRequestTerminated
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
