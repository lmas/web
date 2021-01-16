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
	if err := handler(c); err != nil {
		switch err := errors.Cause(err).(type) {
		case *web.HTTPError:
			http.Error(rec, err.Error(), err.Status())
		default:
			http.Error(rec, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	return rec.Result()
}
