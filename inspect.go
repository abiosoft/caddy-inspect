package inspect

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

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
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	return nil
}

type Details struct {
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
	Cookies []*http.Cookie `json:"cookies"`
	Modules []string       `json:"modules"`
}

func detailsFromRequest(r *http.Request) (d Details) {
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

	for _, m := range caddy.ActiveContext().Modules() {
		d.Modules = append(d.Modules, m.CaddyModule().String())
	}

	return
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.logger.Log(zap.InfoLevel, "the request is being inspected")
	time.Sleep(1 * time.Second)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.Encode(detailsFromRequest(r))

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
