package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web"
	"github.com/pkg/errors"
)

func doRequest(t *testing.T, handler web.Handler, method, path string, headers http.Header, body io.Reader) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if headers != nil {
		req.Header = headers
	}
	c := &web.Context{nil, rec, req, nil}
	if e := handler(c); e != nil {
		switch err := errors.Cause(e).(type) {
		case *web.ErrorHTTP:
			http.Error(rec, err.Error(), err.Status())
		case *web.ErrorPanic:
			http.Error(rec, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		default:
			http.Error(rec, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	return rec.Result()
}
