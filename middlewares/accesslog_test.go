package middlewares

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web/internal/assert"
)

func TestAccessLog(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	var buf bytes.Buffer
	mw := AccessLog(log.New(&buf, "", 0))
	wrapped := mw(h)

	resp := assert.DoRequest(t, wrapped, "GET", "/", nil, nil)
	assert.StatusCode(t, resp, http.StatusOK)
	if len(buf.String()) < 1 {
		t.Errorf("didn't get a log line")
	}
}

func BenchmarkAccessLog(b *testing.B) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200) // Do nothing else
	})
	var buf bytes.Buffer
	mw := AccessLog(log.New(&buf, "", 0))
	wrapped := mw(h)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped.ServeHTTP(w, r)
	}
}
