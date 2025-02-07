package inspect

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
)

type Response struct {
	URL             string      `json:"url,omitempty"`
	Host            string      `json:"host,omitempty"`
	Method          string      `json:"method,omitempty"`
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
	Cookies       []*http.Cookie `json:"cookies,omitempty"`
	ActiveModules []string       `json:"active_modules,omitempty"`
	LoadedModules []string       `json:"loaded_modules,omitempty"`
	responseMode  bool
}

func buildResponse(ctx caddy.Context, w http.ResponseWriter, r *http.Request) (d Response) {
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

	d.LoadedModules = caddy.Modules()
	for _, m := range ctx.Modules() {
		d.ActiveModules = append(d.ActiveModules, m.CaddyModule().String())
	}

	return
}
