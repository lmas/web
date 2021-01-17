package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web"
)

var basicHandler = web.Handler(func(c *web.Context) error {
	return c.String(200, "ok")
})

var benchHandler = web.Handler(func(c *web.Context) error {
	return c.Empty(200) // Do nothing else
})

func doRequest(t *testing.T, handler web.Handler, method, path string, headers http.Header, body io.Reader) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if headers != nil {
		req.Header = headers
	}
	c := &web.Context{nil, rec, req, nil}
	if err := handler(c); err != nil {
		_ = web.SimpleErrorHandler(c, err)
	}
	return rec.Result()
}
