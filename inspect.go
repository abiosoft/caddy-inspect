package inspect

import (
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

	m.server = &Server{logger: m.logger, clear: make(chan struct{})}

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

type Request struct {
	URL           string      `json:"url"`
	Host          string      `json:"host"`
	Method        string      `json:"method"`
	Headers       http.Header `json:"headers"`
	RemoteAddress string      `json:"remote_address"`
	Form          string      `json:"form"`
	Proto         string      `json:"proto"`
	UserAgent     string      `json:"user_agent"`
	Referer       string      `json:"referer"`
	ContentLength int64       `json:"content_length"`
	BasicAuth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"basic_auth"`
	Cookies       []*http.Cookie `json:"cookies"`
	ActiveModules []string       `json:"active_modules"`
	LoadedModules []string       `json:"loaded_modules"`
}

func (m *Middleware) fromHttpRequest(r *http.Request) (d Request) {
	d.URL = r.URL.String()
	d.Method = r.Method
	d.Host = r.Host
	d.Headers = r.Header
	d.RemoteAddress = r.RemoteAddr
	d.Form = r.Form.Encode()
	d.Proto = r.Proto
	d.UserAgent = r.UserAgent()
	d.Referer = r.Referer()
	d.BasicAuth.Username, d.BasicAuth.Password, _ = r.BasicAuth()
	d.ContentLength = r.ContentLength
	d.Cookies = r.Cookies()

	d.LoadedModules = caddy.Modules()
	for _, m := range m.ctx.Modules() {
		d.ActiveModules = append(d.ActiveModules, m.CaddyModule().String())
	}

	return
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.logger.Info("a request is being inspected")

	request := m.fromHttpRequest(r)
	m.server.handle(request)

	return next.ServeHTTP(w, r)
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
