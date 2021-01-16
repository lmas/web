package middlewares

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web"
	"github.com/lmas/web/internal/assert"
)

func TestAccessLog(t *testing.T) {
	h := web.Handler(func(c *web.Context) error {
		return c.String(200, "ok")
	})
	var buf bytes.Buffer
	mw := AccessLog(log.New(&buf, "", 0))
	wrapped := mw(h)

	resp := doRequest(t, wrapped, "GET", "/", nil, nil)
	assert.StatusCode(t, resp, http.StatusOK)
	if len(buf.String()) < 1 {
		t.Errorf("didn't get a log line")
	}
}

func BenchmarkAccessLog(b *testing.B) {
	h := web.Handler(func(c *web.Context) error {
		return c.Empty(200) // Do nothing else
	})
	var buf bytes.Buffer
	mw := AccessLog(log.New(&buf, "", 0))
	wrapped := mw(h)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	c := &web.Context{nil, w, r, nil}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped(c)
	}
}
