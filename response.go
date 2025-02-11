package inspect

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type Response struct {
	URL             string      `json:"url,omitempty"`
	Method          string      `json:"method,omitempty"`
	Host            string      `json:"host,omitempty"`
	RequestHeaders  http.Header `json:"request_headers,omitempty"`
	ResponseHeaders http.Header `json:"response_headers,omitempty"`
	RemoteAddress   string      `json:"remote_address,omitempty"`
	Form            string      `json:"form,omitempty"`
	Proto           string      `json:"proto,omitempty"`
	UserAgent       string      `json:"user_agent,omitempty"`
	Referer         string      `json:"referer,omitempty"`
	ContentLength   int64       `json:"content_length,omitempty"`
	BasicAuth       *struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"basic_auth,omitempty"`
	Cookies      []*http.Cookie `json:"cookies,omitempty"`
	CaddyVersion string         `json:"caddy_version,omitempty"`
	CaddyContext struct {
		Variables map[string]any `json:"Variables,omitempty"`
		Modules   []string       `json:"Modules,omitempty"`
		Error     any            `json:"Error,omitempty"`
	} `json:"caddy_context,omitempty"`
	CaddyModules []string `json:"caddy_modules,omitempty"`
	Caddyfile    *struct {
		File            string   `json:"file,omitempty"`
		Line            int      `json:"line,omitempty"`
		Source          []string `json:"source,omitempty"`
		SourceLineStart int      `json:"source_line_start,omitempty"`
	} `json:"caddyfile,omitempty"`
	responseMode bool
}

func buildResponse(m Middleware, w http.ResponseWriter, r *http.Request) (d Response) {
	d.URL = r.URL.String()
	d.Method = r.Method
	d.Host = r.Host
	d.RequestHeaders = r.Header
	d.RemoteAddress = r.RemoteAddr
	d.Form = r.Form.Encode()
	d.Proto = r.Proto
	d.UserAgent = r.UserAgent()
	d.Referer = r.Referer()
	d.ContentLength = r.ContentLength
	d.Cookies = r.Cookies()

	if w != nil {
		d.ResponseHeaders = w.Header()
		d.responseMode = true
	}

	username, password, _ := r.BasicAuth()
	if username != "" || password != "" {
		d.BasicAuth = &struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
		}{Username: username, Password: password}
	}

	d.CaddyModules = caddy.Modules()
	for _, m := range m.ctx.Modules() {
		d.CaddyContext.Modules = append(d.CaddyContext.Modules, m.CaddyModule().String())
	}

	_, d.CaddyVersion = caddy.Version()

	if m.File != "" && m.Line > 0 {
		d.Caddyfile = &struct {
			File            string   `json:"file,omitempty"`
			Line            int      `json:"line,omitempty"`
			Source          []string `json:"source,omitempty"`
			SourceLineStart int      `json:"source_line_start,omitempty"`
		}{
			File:            m.File,
			Line:            m.Line,
			Source:          m.Source,
			SourceLineStart: m.SourceLineStart,
		}
	}

	vars, _ := r.Context().Value(caddyhttp.VarsCtxKey).(map[string]any)
	d.CaddyContext.Variables = vars

	if err, _ := r.Context().Value(caddyhttp.ErrorCtxKey).(error); err != nil {
		d.CaddyContext.Error = err.Error()

		// if it is an handler error, set specific error
		if err, ok := err.(caddyhttp.HandlerError); ok {
			herr := handlerErr{HandlerError: err}
			if err := err.Err; err != nil {
				herr.Err = err.Error()
			}
			d.CaddyContext.Error = herr
		}
	}

	return
}

type handlerErr struct {
	Err string
	caddyhttp.HandlerError
}
