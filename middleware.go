package inspect

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("inspect", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("inspect", httpcaddyfile.After, "encode")
}

var errRequestTerminated = errors.New("request terminated")

// Middleware implements an HTTP handler that
// inspects the current request.
type Middleware struct {
	logger *zap.Logger
	ctx    caddy.Context

	Key string
	snippetDetails
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
	m.snippetDetails = configMap.get(m.Key)

	if err := setUpServer(ctx); err != nil {
		return err
	}

	server := getServerInstance()
	port, started, err := server.start()
	if err != nil {
		return fmt.Errorf("error occured during provision: %w", err)
	}

	// print the server listen address if not previously running
	if !started {
		m.logger.Info(fmt.Sprintf("inspect console available at http://127.0.0.1:%d", port))
	}

	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	logger := m.logger.With(zap.String("file", m.file), zap.Int("line", m.line))

	logger.Info("inspecting")

	server := getServerInstance()
	action := server.handle(m, nil, r)

	switch action {
	case requestActionResume:
		logger.Info("resumed")
		return next.ServeHTTP(w, r)

	case requestActionStep:
		logger.Info("proceeding to response")

		// process middleware chain
		err := next.ServeHTTP(w, r)

		// handle the updated request details
		action := server.handle(m, w, r)
		if err != nil {
			return err
		}
		if action == requestActionResume {
			return nil
		}

		// request stopped
		fallthrough
	case requestActionStop:
		logger.Info("stopped")
	}

	return caddyhttp.Error(http.StatusServiceUnavailable, errRequestTerminated)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	// persist the file name and line number
	var s snippetDetails
	s.file = d.File()
	s.line = d.Line()

	m.Key = configKey(s.file, s.line)

	if s.file != "" {
		if src, lineStart, err := loadCaddyfileSnippet(s.file, s.line); err == nil {
			s.source = src
			s.sourceLineStart = lineStart
		}

		configMap.set(m.Key, s)
	}

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// loadCaddyfileSnippet reads 5 lines within context in the loaded Caddyfile.
func loadCaddyfileSnippet(file string, line int) (snippet []string, firstLine int, err error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading Caddyfile: %w", err)
	}

	lines := strings.Split(string(f), "\n")
	if len(lines) < line {
		return nil, 0, fmt.Errorf("invalid line: %d", line)
	}

	lineIndex := line - 1 // slice index

	start := lineIndex - 2
	if start < 0 {
		start = 0
	}

	end := (lineIndex + 1) + 2 // the extra 1 is for ending index which is not inclusive.
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end], start + 1, nil
}

type snippetDetails struct {
	file            string
	line            int
	source          []string
	sourceLineStart int // for lack of a better name
}

func (c snippetDetails) valid() bool {
	return c.file != "" && c.line > 0
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
